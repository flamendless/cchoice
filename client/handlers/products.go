package handlers

import (
	"cchoice/client/components"
	"cchoice/client/components/layout"
	"cchoice/internal/logs"
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
	GetProducts() (*pb.ProductsResponse, error)
	GetProductsWithSorting(string, sortDir string) ([]*pb.Product, error)
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

func (h *ProductHandler) ProductTablePage(w http.ResponseWriter, r *http.Request) *HandlerRes {
	res, err := h.ProductService.GetProducts()
	if err != nil {
		return &HandlerRes{Error: err, StatusCode: http.StatusInternalServerError}
	}

	return &HandlerRes{
		Component: layout.Base("Products", components.ProductTableView(res)),
	}
}

func (h *ProductHandler) ProductTableBody(w http.ResponseWriter, r *http.Request) *HandlerRes {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return &HandlerRes{Error: err, StatusCode: http.StatusBadRequest}
	}

	paramSortField := q.Get("sort_field")
	paramSortDir := q.Get("sort_dir")
	logs.Log().Debug(
		"ProductTableBody params",
		zap.String("sort field", paramSortField),
		zap.String("sort dir", paramSortDir),
	)
	if paramSortField == "" || paramSortDir == "" {
		return &HandlerRes{Error: errors.New("Invalid URL params"), StatusCode: http.StatusBadRequest}
	}

	res, err := h.ProductService.GetProductsWithSorting(paramSortField, paramSortDir)
	if err != nil {
		return &HandlerRes{Error: err, StatusCode: http.StatusInternalServerError}
	}

	return &HandlerRes{
		Component: components.ProductTableBody(res),
	}
}
