package handlers

import (
	"cchoice/client/components"
	"cchoice/client/components/layout"
	"cchoice/client/services"
	"cchoice/internal/logs"
	"net/http"

	"go.uber.org/zap"
)

type ProductHandler struct {
	Logger         *zap.Logger
	ProductService *services.ProductService
}

func NewProductHandler(
	logger *zap.Logger,
	productService *services.ProductService,
) ProductHandler {
	return ProductHandler{
		Logger:         logger,
		ProductService: productService,
	}
}

func (h *ProductHandler) ProductTablePage(w http.ResponseWriter, r *http.Request) {
	res, err := h.ProductService.GetProducts()
	if err != nil {
		logs.LogHTTPHandler(h.Logger, r, err)

		//error page
		return
	}

	layout.Base(
		"Products",
		components.ProductTableView(res),
	).Render(r.Context(), w)
}

func (h *ProductHandler) ProductTableBody(w http.ResponseWriter, r *http.Request) {
	paramSortField := r.URL.Query().Get("sort_field")
	paramSortDir := r.URL.Query().Get("sort_dir")
	logs.Log().Info(
		"products_table params",
		zap.String("sort field", paramSortField),
		zap.String("sort dir", paramSortDir),
	)
	if paramSortField == "" || paramSortDir == "" {

		//error page
		return
	}

	res, err := h.ProductService.GetProductsWithSorting(paramSortField, paramSortDir)
	if err != nil {
		logs.LogHTTPHandler(h.Logger, r, err)

		//error page
		return
	}

	components.ProductTableBody(res).Render(r.Context(), w)
}
