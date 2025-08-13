package cmd

import (
	"bytes"
	"cchoice/cmd/parse_map/enums"
	maps_models "cchoice/cmd/parse_map/models"
	products_models "cchoice/cmd/parse_products/models"
	"cchoice/internal/logs"
	"fmt"
	"go/format"
	"os"
	"text/template"
	"time"

	"github.com/goccy/go-json"
	"github.com/gookit/goutil/dump"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ParseMapFlags struct {
	filepath string
	jsonOut  bool
}

var metrics *products_models.Metrics
var cmdParseMapFlags ParseMapFlags

const (
	ROW_ID       = 0
	ROW_NAME     = 1
	ROW_CODE     = 2
	ROW_LEVEL    = 3
	ROW_OLD_NAME = 4
)

func init() {
	f := cmdParseMap.Flags
	f().StringVarP(&cmdParseMapFlags.filepath, "filepath", "p", "", "Filepath to the XLSX file")
	f().BoolVarP(&cmdParseMapFlags.jsonOut, "json", "", false, "Output to JSON (stdout)")
	if err := cmdParseMap.MarkFlagRequired("filepath"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmdParseMap)
}

var cmdParseMap = &cobra.Command{
	Use:   "parse_map",
	Short: "Parse PH dataset to map structure",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Parse map",
			zap.String("filepath", cmdParseMapFlags.filepath),
			zap.Bool("json", cmdParseMapFlags.jsonOut),
		)

		const SHEET_NAME = "PSGC"

		file, err := excelize.OpenFile(cmdParseMapFlags.filepath)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				logs.Log().Error(err.Error())
				return
			}
		}()

		processTime := time.Now()
		metrics = &products_models.Metrics{}
		defer func() {
			metrics.Add("process", time.Since(processTime))
			metrics.LogTime(logs.Log())
		}()

		rows, err := file.Rows(SHEET_NAME)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := rows.Close(); err != nil {
				logs.Log().Error(err.Error())
				return
			}
		}()

		rowIdx := 0
		skipInitialRows := 0
		for range skipInitialRows + 1 {
			rows.Next()
			rowIdx++
		}

		const LEN_ROOT = 512
		const LEN_SUB = 256

		data := make([]*maps_models.Map, 0, LEN_ROOT)
		var current *maps_models.Map

		for rows.Next() {
			rowIdx++
			row, err := rows.Columns()
			if err != nil {
				logs.Log().Info(
					"Error processing row",
					zap.Error(err),
					zap.Int("idx", rowIdx),
				)
				continue
			}

			if len(row) == 0 {
				break
			}

			name := row[ROW_NAME]
			oldName := row[ROW_OLD_NAME]
			if oldName != "" {
				name = fmt.Sprintf("%s (%s)", name, oldName)
			}

			level := enums.ParseXLSXLevelToEnum(row[ROW_LEVEL])
			if level == enums.LEVEL_UNDEFINED {
				continue
			}

			for current != nil && !isParentLevel(level, current.Level) {
				current = current.Parent
			}

			newMap := &maps_models.Map{
				ID:       row[ROW_ID],
				Name:     name,
				Code:     row[ROW_CODE],
				Level:    level,
				Contents: make([]*maps_models.Map, 0, LEN_SUB),
				Parent:   current,
			}
			if current == nil {
				data = append(data, newMap)
				current = newMap
			} else {
				current.Contents = append(current.Contents, newMap)
				if level != enums.LEVEL_BARANGAY {
					current = newMap
				}
			}
		}

		// dump.Println(data)
		dump.Println("Done parsing map data")

		processTimeSort := time.Now()
		maps_models.SortMap(data)
		metrics.Add("sorting", time.Since(processTimeSort))

		tmpl, err := template.ParseFiles("./cmd/parse_map/templates/map.go.tmpl")
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			panic(err)
		}

		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			panic(err)
		}

		if err := os.WriteFile("./cmd/parse_map/models/maps_generated.go", formatted, 0644); err != nil {
			panic(err)
		}

		if cmdParseMapFlags.jsonOut {
			jsonData, err := json.MarshalIndent(data, "", "    ")
			if err != nil {
				panic(err)
			}
			if err := os.WriteFile("./tmp/generated_map.json", jsonData, 0644); err != nil {
				panic(err)
			}
		}
	},
}

func isParentLevel(child enums.Level, parent enums.Level) bool {
	switch child {
	case enums.LEVEL_REGION:
		return parent == enums.LEVEL_UNDEFINED // root only
	case enums.LEVEL_PROVINCE:
		return parent == enums.LEVEL_REGION
	case enums.LEVEL_CITY, enums.LEVEL_MUNICIPALITY:
		return parent == enums.LEVEL_PROVINCE
	case enums.LEVEL_BARANGAY:
		return parent == enums.LEVEL_CITY || parent == enums.LEVEL_MUNICIPALITY
	default:
		return false
	}
}
