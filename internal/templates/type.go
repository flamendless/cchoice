package templates

import (
	"cchoice/internal/models"
	"fmt"
	"strings"
)

type Column struct {
	Index    int
	Required bool
}

type RowToProductHandler (func(*Template, []string) (*models.Product, error))

type Template struct {
	Columns      map[string]*Column
	RowToProduct RowToProductHandler
}

func CreateTemplate(kind TemplateKind) *Template {
	switch kind {
	case Undefined:
		panic("Can't use undefined template")
	case Sample:
		return &Template{
			Columns:      SampleColumns,
			RowToProduct: SampleRowToProduct,
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
