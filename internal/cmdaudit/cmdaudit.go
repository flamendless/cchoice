package cmdaudit

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/logs"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const (
	EnvCLIStaffID       = "CCHOICE_CLI_STAFF_ID"
	SystemCLIStaffEmail = "cli@system.internal"
	maxResultErrorLen   = 500
)

var skipCommands = map[string]struct{}{
	"web":         {},
	"api":         {},
	"datamigrate": {},
}

type auditResult struct {
	Status     string         `json:"status"`
	Command    string         `json:"command"`
	Args       []string       `json:"args,omitempty"`
	Flags      map[string]any `json:"flags,omitempty"`
	DurationMS int64          `json:"duration_ms"`
	Summary    string         `json:"summary,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	Error      string         `json:"error,omitempty"`
}

type resultData struct {
	summary string
	details map[string]any
}

type resultKey struct{}

var (
	staffIDOnce sync.Once
	cachedStaff int64
	staffIDErr  error
)

func ShouldSkipCommand(cmd *cobra.Command) bool {
	if cmd == nil {
		return true
	}
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "" {
			continue
		}
		if _, skip := skipCommands[c.Name()]; skip {
			return true
		}
	}
	return cmd.Name() == ""
}

func isSystemCLIRun() bool {
	return strings.TrimSpace(os.Getenv(EnvCLIStaffID)) == ""
}

func SetSummary(ctx context.Context, summary string) context.Context {
	data := resultFromContext(ctx)
	data.summary = summary
	return context.WithValue(ctx, resultKey{}, data)
}

func SetDetails(ctx context.Context, details map[string]any) context.Context {
	data := resultFromContext(ctx)
	data.details = details
	return context.WithValue(ctx, resultKey{}, data)
}

func Log(ctx context.Context, db database.IService, encoder encode.IEncode, cmd *cobra.Command, args []string, runErr error, duration time.Duration) {
	if cmd == nil || ShouldSkipCommand(cmd) || isSystemCLIRun() {
		return
	}

	staffID, err := resolveStaffID(ctx, db, encoder)
	if err != nil {
		logs.Log().Warn(
			"[CmdAudit] failed to resolve staff ID, skipping audit log",
			zap.String("command", cmd.Name()),
			zap.Error(err),
		)
		return
	}

	resultJSON, err := buildResult(cmd, args, runErr, duration, resultFromContext(ctx))
	if err != nil {
		logs.Log().Warn(
			"[CmdAudit] failed to build audit result",
			zap.String("command", cmd.Name()),
			zap.Error(err),
		)
		return
	}

	if _, err := db.GetQueries().CreateStaffLog(ctx, queries.CreateStaffLogParams{
		StaffID:     staffID,
		Action:      constants.ActionRun,
		Module:      cmd.Name(),
		Result:      resultJSON,
		UseragentID: sql.NullInt64{},
	}); err != nil {
		logs.Log().Warn(
			"[CmdAudit] failed to create staff log",
			zap.String("command", cmd.Name()),
			zap.Error(err),
		)
	}
}

func resolveStaffID(ctx context.Context, db database.IService, encoder encode.IEncode) (int64, error) {
	staffIDOnce.Do(func() {
		if encoded := strings.TrimSpace(os.Getenv(EnvCLIStaffID)); encoded != "" {
			decoded := encoder.Decode(encoded)
			if decoded == encode.INVALID {
				staffIDErr = errInvalidCLIStaffID
				return
			}
			cachedStaff = decoded
			return
		}

		staff, err := db.GetQueries().GetStaffByEmail(ctx, SystemCLIStaffEmail)
		if err != nil {
			staffIDErr = err
			return
		}
		cachedStaff = staff.ID
	})

	return cachedStaff, staffIDErr
}

func buildResult(cmd *cobra.Command, args []string, runErr error, duration time.Duration, extra *resultData) (string, error) {
	result := auditResult{
		Command:    cmd.Name(),
		Args:       args,
		Flags:      collectSanitizedFlags(cmd),
		DurationMS: duration.Milliseconds(),
		Status:     "success",
	}

	if runErr != nil {
		result.Status = "error"
		result.Error = truncateError(runErr.Error())
	}

	if extra != nil {
		result.Summary = extra.summary
		if len(extra.details) > 0 {
			result.Details = extra.details
		}
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func collectSanitizedFlags(cmd *cobra.Command) map[string]any {
	flags := make(map[string]any)
	if cmd == nil {
		return flags
	}

	visit := func(set *pflag.FlagSet) {
		if set == nil {
			return
		}
		set.VisitAll(func(f *pflag.Flag) {
			if f == nil || shouldRedactFlag(f.Name) {
				return
			}
			if !f.Changed {
				return
			}
			flags[f.Name] = f.Value.String()
		})
	}

	visit(cmd.Flags())
	visit(cmd.PersistentFlags())
	return flags
}

func shouldRedactFlag(name string) bool {
	lower := strings.ToLower(name)
	for _, part := range []string{"password", "secret", "token", "key"} {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}

func truncateError(msg string) string {
	if len(msg) <= maxResultErrorLen {
		return msg
	}
	return strings.TrimRightFunc(msg[:maxResultErrorLen], func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) + "..."
}

func resultFromContext(ctx context.Context) *resultData {
	if ctx == nil {
		return &resultData{}
	}
	if data, ok := ctx.Value(resultKey{}).(*resultData); ok && data != nil {
		return data
	}
	return &resultData{}
}

// ResetStaffIDCache clears cached staff ID resolution (for tests).
func ResetStaffIDCache() {
	staffIDOnce = sync.Once{}
	cachedStaff = 0
	staffIDErr = nil
}

var errInvalidCLIStaffID = &invalidCLIStaffIDError{}

type invalidCLIStaffIDError struct{}

func (e *invalidCLIStaffIDError) Error() string {
	return "invalid CCHOICE_CLI_STAFF_ID"
}
