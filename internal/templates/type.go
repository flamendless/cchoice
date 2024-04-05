package templates

import (
	"cchoice/internal"
	"cchoice/internal/models"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Column struct {
	Index    int
	Required bool
}

type Template struct {
	AppContext       *internal.AppContext
	SkipInitialRows  int
	AssumedRowsCount int
	Columns          map[string]*Column
	RowToProduct     func(*Template, []string) (*models.Product, []error)
	ProcessRows      func(*Template, *excelize.Rows) []*models.Product
}

func CreateTemplate(kind TemplateKind) *Template {
	switch kind {
	case Undefined:
		panic("Can't use undefined template")
	case Sample:
		return &Template{
			SkipInitialRows:  0,
			AssumedRowsCount: 128,
			Columns:          SampleColumns,
			RowToProduct:     SampleRowToProduct,
			ProcessRows:      SampleProcessRows,
		}
	case DeltaPlus:
		return &Template{
			SkipInitialRows:  1,
			AssumedRowsCount: 1024,
			Columns:          DeltaPlusColumns,
			RowToProduct:     DeltaPlusRowToProduct,
			ProcessRows:      DeltaPlusProcessRows,
		}
	}
	return nil
}

func (tpl *Template) Print() {
	for k, v := range tpl.Columns {
		fmt.Printf("%s = %d\n", k, v.Index)
	}
}

func (tpl *Template) AlignRow(row []string) []string {
	if len(row) >= len(tpl.Columns) {
		return row
	}

	res := make([]string, len(tpl.Columns))
	for i := 1; i < len(row); i++ {
		res[i] = row[i]
	}
	return res
}

func (tpl *Template) ValidateColumns() bool {
	var result bool = true
	notFoundColumns := make([]string, 0, len(tpl.Columns))

	for key, col := range tpl.Columns {
		if col.Required && col.Index == -1 {
			notFoundColumns = append(notFoundColumns, key)
			result = false
		}
	}

	if !result {
		fmt.Printf(
			"Failed to find index for required column(s) for '%s'\n",
			strings.Join(notFoundColumns, ", "),
		)
	}

	return result
}

var _ ITemplate = (*Template)(nil)
