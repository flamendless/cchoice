package cmd

import (
	"time"

	"cchoice/internal/cmdaudit"
	"cchoice/internal/database"
	"cchoice/internal/encode/sqids"

	"github.com/spf13/cobra"
)

var commandAuditEnabled bool

func EnableCommandAudit(root *cobra.Command) {
	if commandAuditEnabled || root == nil {
		return
	}
	commandAuditEnabled = true
	walkCommands(root, wrapCommandAudit)
}

func walkCommands(cmd *cobra.Command, fn func(*cobra.Command)) {
	for _, sub := range cmd.Commands() {
		if sub == nil {
			continue
		}
		fn(sub)
		walkCommands(sub, fn)
	}
}

func wrapCommandAudit(cmd *cobra.Command) {
	if cmd == nil || cmd.Name() == "" || cmdaudit.ShouldSkipCommand(cmd.Name()) {
		return
	}

	if cmd.RunE != nil {
		orig := cmd.RunE
		cmd.RunE = func(c *cobra.Command, args []string) error {
			start := time.Now()
			err := orig(c, args)
			logCommandAudit(c, args, err, time.Since(start))
			return err
		}
		return
	}

	if cmd.Run != nil {
		orig := cmd.Run
		cmd.Run = func(c *cobra.Command, args []string) {
			start := time.Now()
			orig(c, args)
			logCommandAudit(c, args, nil, time.Since(start))
		}
	}
}

func logCommandAudit(cmd *cobra.Command, args []string, runErr error, duration time.Duration) {
	if cmd == nil || cmdaudit.ShouldSkipCommand(cmd.Name()) {
		return
	}

	db := database.New(database.DB_MODE_RW)
	defer db.Close()

	encoder := sqids.MustSqids()
	cmdaudit.Log(cmd.Context(), db, encoder, cmd, args, runErr, duration)
}
