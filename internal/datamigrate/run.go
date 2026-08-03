package datamigrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/database"
)

type CommandRunner func(ctx context.Context, argv []string) error

func DefaultCommandRunner() CommandRunner {
	return func(ctx context.Context, argv []string) error {
		exe, prefix, err := resolveCLICommand()
		if err != nil {
			return err
		}
		cmd := exec.CommandContext(ctx, exe, append(prefix, argv...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
}

// resolveCLICommand returns the binary and optional prefix args (e.g. go run) for child CLI scripts.
// os.Executable() is unreliable under go run; fall back to go run ./main.go with build tags.
func resolveCLICommand() (exe string, prefix []string, err error) {
	if bin := strings.TrimSpace(os.Getenv("CCHOICE_BIN")); bin != "" {
		return bin, nil, nil
	}

	exe, err = os.Executable()
	if err != nil {
		return "", nil, err
	}
	if filepath.Base(exe) == "go" {
		tags := strings.TrimSpace(os.Getenv("CCHOICE_BUILD_TAGS"))
		if tags == "" {
			tags = "fts5,staticfs"
		}
		return exe, []string{"run", "-tags=" + tags, "./main.go"}, nil
	}
	return exe, nil, nil
}

func Up(ctx context.Context, db database.IService, runner CommandRunner) error {
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
	if runner == nil {
		runner = DefaultCommandRunner()
	}

	for _, pending := range state.PendingScripts {
		script, ok := scriptByName(pending.Name)
		if !ok {
			return errors.New(constants.MsgUnknownDatamigrateScript + pending.Name)
		}

		fmt.Fprintf(os.Stdout, "running post-migrate script: %s\n", pending.Name)
		if err := runner(ctx, script.BuildRunArgs()); err != nil {
			return fmt.Errorf("post-migrate script %s failed: %w", pending.Name, err)
		}

		state, err = GetState(ctx, db)
		if err != nil {
			return err
		}
		for _, p := range state.PendingScripts {
			if p.Name == pending.Name {
				return fmt.Errorf("post-migrate script %s did not record as applied", pending.Name)
			}
		}
	}

	return nil
}

func scriptByName(name string) (Script, bool) {
	for i := range Scripts {
		if Scripts[i].Name == name {
			return Scripts[i], true
		}
	}
	return Script{}, false
}
