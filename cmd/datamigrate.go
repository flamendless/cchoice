package cmd

import (
	"fmt"
	"os"

	"cchoice/internal/database"
	"cchoice/internal/datamigrate"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdDatamigrate)
	cmdDatamigrate.AddCommand(cmdDatamigrateCheck)
	cmdDatamigrate.AddCommand(cmdDatamigrateStatus)
	cmdDatamigrate.AddCommand(cmdDatamigrateUp)
	cmdDatamigrate.AddCommand(cmdDatamigrateDoctor)
	cmdDatamigrate.AddCommand(cmdDatamigrateMark)
}

var cmdDatamigrate = &cobra.Command{
	Use:   "datamigrate",
	Short: "Post-migrate CLI script tracking",
}

var cmdDatamigrateCheck = &cobra.Command{
	Use:   "check",
	Short: "Fail if required post-migrate scripts have not been applied",
	Run: func(cmd *cobra.Command, args []string) {
		db := database.New(database.DB_MODE_RO)
		defer db.Close()

		if err := datamigrate.Check(cmd.Context(), db); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

var cmdDatamigrateStatus = &cobra.Command{
	Use:   "status",
	Short: "Show applied and pending post-migrate scripts",
	Run: func(cmd *cobra.Command, args []string) {
		db := database.New(database.DB_MODE_RO)
		defer db.Close()

		ctx := cmd.Context()

		pendingGoose, err := datamigrate.ListPendingGooseMigrations(ctx, db)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		version, err := datamigrate.GooseMaxAppliedVersion(ctx, db)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		state, err := datamigrate.GetState(ctx, db)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Println(datamigrate.FormatStatus(state, version, pendingGoose))
	},
}

var cmdDatamigrateUp = &cobra.Command{
	Use:   "up",
	Short: "Run pending post-migrate scripts in order",
	Run: func(cmd *cobra.Command, args []string) {
		db := database.New(database.DB_MODE_RW)
		defer db.Close()

		if err := datamigrate.Up(cmd.Context(), db, datamigrate.DefaultCommandRunner()); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

var cmdDatamigrateDoctor = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose common migration and post-migrate data issues",
	Run: func(cmd *cobra.Command, args []string) {
		db := database.New(database.DB_MODE_RO)
		defer db.Close()

		findings, err := datamigrate.Doctor(cmd.Context(), db)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Println(datamigrate.FormatDoctor(findings))
		for _, f := range findings {
			if f.Severity == "error" {
				os.Exit(1)
			}
		}
	},
}

var cmdDatamigrateMark = &cobra.Command{
	Use:   "mark [name]",
	Short: "Manually mark a post-migrate script as applied (break-glass)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db := database.New(database.DB_MODE_RW)
		defer db.Close()

		if err := datamigrate.MarkApplied(cmd.Context(), db, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}
