package datamigrate

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type PendingScript struct {
	Order        int
	AfterVersion int64
	Name         string
	RunHint      string
}

type State struct {
	PendingMigration bool
	PendingScripts   []PendingScript
}

func Check(ctx context.Context, db database.IService) error {
	state, err := GetState(ctx, db)
	if err != nil {
		return err
	}
	if state.PendingMigration {
		return constants.ErrPendingDatamigrateMigration
	}
	if len(state.PendingScripts) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("there are pending post-migrate scripts (run in order):\n")
	for _, p := range state.PendingScripts {
		b.WriteString("  ")
		b.WriteString(formatPendingScriptLine(p.Order, p))
		b.WriteByte('\n')
	}
	return errors.New(strings.TrimSuffix(b.String(), "\n"))
}

func GetState(ctx context.Context, db database.IService) (*State, error) {
	q := db.GetQueries()

	exists, err := datamigrateTableExists(ctx, db)
	if err != nil {
		return nil, errors.New(constants.MsgFailedCheckDatamigrateTable + ": " + err.Error())
	}
	if !exists {
		return &State{PendingMigration: true}, nil
	}

	currentVersion, err := GooseMaxAppliedVersion(ctx, db)
	if err != nil {
		return nil, errors.New(constants.MsgFailedReadGooseVersion + ": " + err.Error())
	}

	appliedRows, err := q.ListDatamigrateAppliedNames(ctx)
	if err != nil {
		return nil, errors.New(constants.MsgFailedListAppliedDatamigrateScripts + ": " + err.Error())
	}

	applied := make(map[string]struct{}, len(appliedRows))
	for _, name := range appliedRows {
		applied[name] = struct{}{}
	}

	pending := make([]PendingScript, 0)
	order := 0
	for _, script := range Scripts {
		if script.AfterVersion > currentVersion {
			continue
		}
		if script.When != nil && !script.When() {
			continue
		}
		if _, ok := applied[script.Name]; ok {
			continue
		}
		order++
		pending = append(pending, PendingScript{
			Order:        order,
			AfterVersion: script.AfterVersion,
			Name:         script.Name,
			RunHint:      script.RunHint,
		})
	}

	return &State{PendingScripts: pending}, nil
}

func TryRecord(ctx context.Context, db database.IService, cmd *cobra.Command, args []string, runErr error) error {
	if runErr != nil || cmd == nil {
		return nil
	}
	if utils.IsDryRun(cmd) {
		return nil
	}

	script, ok := MatchScript(cmd, args)
	if !ok {
		return nil
	}

	exists, err := datamigrateTableExists(ctx, db)
	if err != nil {
		return errors.New(constants.MsgFailedCheckDatamigrateTable + ": " + err.Error())
	}
	if !exists {
		return nil
	}

	if err := db.GetQueries().UpsertDatamigrateApplied(ctx, script.Name); err != nil {
		return errors.New(constants.MsgFailedRecordDatamigrateScript + script.Name + ": " + err.Error())
	}
	return nil
}

func MatchScript(cmd *cobra.Command, args []string) (*Script, bool) {
	commandName := cmd.Name()
	for i := range Scripts {
		script := Scripts[i]
		if script.Command != commandName {
			continue
		}
		if !matchesRequiredArgs(cmd, args, script.RequiredArgs) {
			continue
		}
		return &Scripts[i], true
	}
	return nil, false
}

func matchesRequiredArgs(cmd *cobra.Command, args []string, required []string) bool {
	if len(required) == 0 {
		return true
	}

	tokens := collectArgTokens(cmd, args)
	for _, req := range required {
		if !containsToken(tokens, req) {
			return false
		}
	}
	return true
}

func collectArgTokens(cmd *cobra.Command, args []string) []string {
	tokens := append([]string{}, args...)
	if cmd == nil {
		return tokens
	}

	visit := func(set *pflag.FlagSet) {
		if set == nil {
			return
		}
		set.VisitAll(func(f *pflag.Flag) {
			val := f.Value.String()
			if val == "" || val == "false" {
				return
			}
			tokens = append(tokens, val)
		})
	}

	visit(cmd.Flags())
	visit(cmd.PersistentFlags())
	return tokens
}

func containsToken(tokens []string, want string) bool {
	return slices.Contains(tokens, want)
}

func FormatStatus(state *State, currentVersion int64, pendingGoose []PendingGooseMigration) string {
	if state == nil {
		return ""
	}

	var b strings.Builder
	if len(pendingGoose) > 0 {
		b.WriteString("pending goose migrations (run mage dbup first):\n")
		for i, migration := range pendingGoose {
			b.WriteString("  ")
			b.WriteString(formatPendingGooseLine(i+1, migration))
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	} else {
		b.WriteString("goose migrations: all applied\n\n")
	}

	fmt.Fprintf(&b, "goose max applied version: %d\n", currentVersion)
	if state.PendingMigration {
		b.WriteString("datamigrate table: missing (run mage dbup)\n")
		return strings.TrimSuffix(b.String(), "\n")
	}

	b.WriteString("datamigrate table: ok\n")

	if len(state.PendingScripts) == 0 {
		b.WriteString("pending post-migrate scripts: none\n")
		return strings.TrimSuffix(b.String(), "\n")
	}

	b.WriteString("pending post-migrate scripts (run in order):\n")
	for _, p := range state.PendingScripts {
		b.WriteString("  ")
		b.WriteString(formatPendingScriptLine(p.Order, p))
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func MarkApplied(ctx context.Context, db database.IService, name string) error {
	for _, script := range Scripts {
		if script.Name == name {
			exists, err := datamigrateTableExists(ctx, db)
			if err != nil {
				return errors.New(constants.MsgFailedCheckDatamigrateTable + ": " + err.Error())
			}
			if !exists {
				return constants.ErrPendingDatamigrateMigration
			}
			if err := db.GetQueries().UpsertDatamigrateApplied(ctx, name); err != nil {
				return errors.New(constants.MsgFailedMarkDatamigrateScript + name + ": " + err.Error())
			}
			return nil
		}
	}
	return errors.New(constants.MsgUnknownDatamigrateScript + name)
}
