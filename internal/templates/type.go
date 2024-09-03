package templates

import (
	"cchoice/internal/ctx"
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
	AppFlags *ctx.ParseXLSXFlags
	CtxApp   *ctx.App
	Brand    *models.Brand

	SkipInitialRows  int
	AssumedRowsCount int
	BrandOnly        bool
	Columns          map[string]*Column
	RowToProduct     func(*Template, []string) (*models.Product, []error)
	RowToSpecs       func(*Template, []string) *models.ProductSpecs
	ProcessRows      func(*Template, *excelize.Rows) []*models.Product
}

func CreateTemplate(kind TemplateKind) *Template {
	switch kind {
	case TEMPLATE_UNDEFINED:
		panic("Can't use undefined template")
	case TEMPLATE_SAMPLE:
		return &Template{
			SkipInitialRows:  0,
			AssumedRowsCount: 128,
			Columns:          SampleColumns,
			RowToProduct:     SampleRowToProduct,
			RowToSpecs:       SampleRowToSpecs,
			ProcessRows:      SampleProcessRows,
		}

	case TEMPLATE_DELTAPLUS:
		return &Template{
			SkipInitialRows:  1,
			AssumedRowsCount: 1024,
			Columns:          DeltaPlusColumns,
			RowToProduct:     DeltaPlusRowToProduct,
			RowToSpecs:       DeltaPlusRowToSpecs,
			ProcessRows:      DeltaPlusProcessRows,
		}

	case TEMPLATE_BOSCH:
		return &Template{
			SkipInitialRows:  2,
			AssumedRowsCount: 256,
			Columns:          BoschColumns,
			RowToProduct:     BoschRowToProduct,
			RowToSpecs:       BoschRowToSpecs,
			ProcessRows:      BoschProcessRows,
		}

	case TEMPLATE_SPARTAN:
		return &Template{BrandOnly: true}
	case TEMPLATE_SHINSETSU:
		return &Template{BrandOnly: true}
	case TEMPLATE_REDMAX:
		return &Template{BrandOnly: true}
	case TEMPLATE_BRADFORD:
		return &Template{BrandOnly: true}
	case TEMPLATE_KOBEWEL:
		return &Template{BrandOnly: true}

	default:
		panic(fmt.Sprintf("No template parser found for %s", kind.String()))
	}
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
