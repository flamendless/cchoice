package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
)

var flagYear int

type HolidayJSON struct {
	Date string `json:"date"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func init() {
	f := cmdSeedHolidays.Flags
	f().IntVarP(&flagYear, "year", "y", time.Now().Year(), "Year to seed holidays for")

	rootCmd.AddCommand(cmdSeedHolidays)
}

var cmdSeedHolidays = &cobra.Command{
	Use:   "seed_holidays",
	Short: "Seed holidays from JSON file for a given year",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		db := database.New(database.DB_MODE_RW)
		defer db.Close()

		assetsDir := filepath.Join("assets", "holidays")
		jsonFile := filepath.Join(assetsDir, fmt.Sprintf("%d.json", flagYear))

		data, err := os.ReadFile(jsonFile)
		if err != nil {
			logs.Log().Error("failed to read holiday JSON file", zap.Error(err), zap.String("file", jsonFile))
			return
		}

		var holidays []HolidayJSON
		if err := json.Unmarshal(data, &holidays); err != nil {
			logs.Log().Error("failed to parse holiday JSON", zap.Error(err))
			return
		}

		logs.Log().Info("seeding holidays", zap.Int("year", flagYear), zap.Int("count", len(holidays)))

		var success int
		q := db.GetQueries()
		for _, h := range holidays {
			holidayType := enums.ParseHolidayTypeToEnum(h.Type)
			if holidayType == enums.HOLIDAY_TYPE_UNDEFINED {
				logs.Log().Warn("invalid holiday type", zap.String("type", h.Type), zap.String("date", h.Date))
				continue
			}

			_, err := q.CreateHoliday(ctx, queries.CreateHolidayParams{
				Date: h.Date,
				Name: h.Name,
				Type: holidayType.String(),
			})
			if err != nil {
				logs.Log().Error("failed to create holiday", zap.Error(err), zap.String("date", h.Date))
				continue
			}

			success++
			logs.Log().Debug("holiday seeded", zap.String("date", h.Date), zap.String("name", h.Name))
		}

		logs.Log().Info(
			"holidays seeded successfully",
			zap.Int("year", flagYear),
			zap.Int("successful", success),
			zap.Int("total", len(holidays)),
		)
	},
}
