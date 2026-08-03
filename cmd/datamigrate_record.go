package cmd

import (
	"cchoice/internal/database"
	"cchoice/internal/datamigrate"
	"cchoice/internal/logs"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func recordDatamigrateSuccess(cmd *cobra.Command, args []string, db database.IService) {
	if err := datamigrate.TryRecord(cmd.Context(), db, cmd, args, nil); err != nil {
		logs.Log().Warn(
			"[Datamigrate] failed to record applied script",
			zap.String("command", cmd.Name()),
			zap.Error(err),
		)
	}
}
