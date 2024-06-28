package handlers

import (
	"cchoice/client/components"
	"cchoice/client/components/layout"
	"cchoice/internal/enums"
	pb "cchoice/proto"
	"errors"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"go.uber.org/zap"
)

type HandlerRes struct {
	Component  templ.Component
	Error      error
	StatusCode int
}

type ProductService interface {
	GetProductsWithSorting(pb.SortField, pb.SortDir) (*pb.ProductsResponse, error)
}

type ProductHandler struct {
	Logger         *zap.Logger
	ProductService ProductService
}

func NewProductHandler(
	logger *zap.Logger,
	productService ProductService,
) ProductHandler {
	return ProductHandler{
		Logger:         logger,
		ProductService: productService,
	}
}

func (h *ProductHandler) ProductTablePage(w *http.ResponseWriter, r *http.Request) *HandlerRes {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return &HandlerRes{Error: err, StatusCode: http.StatusBadRequest}
	}

	sortField := pb.SortField_NAME
	sortDir := pb.SortDir_ASC
	qSortField := q.Get("sort_field")
	qSortDir := q.Get("sort_dir")

	if qSortField != "" || qSortDir != "" {
		sortField = enums.ParseSortFieldEnumPB(qSortField)
		sortDir = enums.ParseSortDirEnumPB(qSortDir)
		if sortField == pb.SortField_SORT_FIELD_UNDEFINED || sortDir == pb.SortDir_SORT_DIR_UNDEFINED {
			return &HandlerRes{Error: errors.New("Invalid URL params"), StatusCode: http.StatusBadRequest}
		}
	}

	res, err := h.ProductService.GetProductsWithSorting(sortField, sortDir)
	if err != nil {
		return &HandlerRes{Error: err, StatusCode: http.StatusInternalServerError}
	}

	return &HandlerRes{
		Component: layout.Base("Products", components.ProductTableView(res)),
	}
}
