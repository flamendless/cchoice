package templates

import (
	"cchoice/cmd/parse_xlsx/models"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Column struct {
	Index    int
	Required bool
}

type Template struct {
	AppFlags *models.ParseXLSXFlags
	CtxApp   *models.ParseXLSX
	Brand    *models.Brand

	Columns               map[string]*Column
	RowToProduct          func(*Template, []string) (*models.Product, []error)
	RowToSpecs            func(*Template, []string) *models.ProductSpecs
	ProcessProductImage   func(*Template, *models.Product) (*models.ProductImage, error)
	ProcessRows           func(*Template, *excelize.Rows) []*models.Product
	GetPromotedCategories func() []string

	SkipInitialRows  int
	AssumedRowsCount int
	BrandOnly        bool
}

func CreateTemplate(kind TemplateKind) *Template {
	switch kind {
	case TEMPLATE_UNDEFINED:
		panic("Can't use undefined template")
	case TEMPLATE_SAMPLE:
		return &Template{
			SkipInitialRows:       0,
			AssumedRowsCount:      128,
			Columns:               SampleColumns,
			RowToProduct:          SampleRowToProduct,
			RowToSpecs:            SampleRowToSpecs,
			ProcessRows:           SampleProcessRows,
			GetPromotedCategories: SampleGetPromotedCategories,
		}

	case TEMPLATE_DELTAPLUS:
		return &Template{
			SkipInitialRows:       1,
			AssumedRowsCount:      1024,
			Columns:               DeltaPlusColumns,
			RowToProduct:          DeltaPlusRowToProduct,
			RowToSpecs:            DeltaPlusRowToSpecs,
			ProcessRows:           DeltaPlusProcessRows,
			GetPromotedCategories: DeltaPlusGetPromotedCategories,
		}

	case TEMPLATE_BOSCH:
		return &Template{
			SkipInitialRows:       2,
			AssumedRowsCount:      256,
			Columns:               BoschColumns,
			RowToProduct:          BoschRowToProduct,
			RowToSpecs:            BoschRowToSpecs,
			ProcessRows:           BoschProcessRows,
			ProcessProductImage:   BoschProcessProductImage,
			GetPromotedCategories: BoschGetPromotedCategories,
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
