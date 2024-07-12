package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	pb "cchoice/proto"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type ProductService interface {
	GetProductsWithSorting(pb.SortField_SortField, pb.SortDir_SortDir) (*pb.ProductsResponse, error)
}

type ProductHandler struct {
	Logger         *zap.Logger
	ProductService ProductService
	AuthService    AuthService
}

func NewProductHandler(
	logger *zap.Logger,
	productService ProductService,
	authService AuthService,
) ProductHandler {
	return ProductHandler{
		Logger:         logger,
		ProductService: productService,
		AuthService:    authService,
	}
}

func (h ProductHandler) ProductTablePage(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	resAuth, _ := h.AuthService.Authenticated(enums.AUD_API, w, r)
	if resAuth != nil {
		return resAuth
	}

	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return &common.HandlerRes{Error: err, StatusCode: http.StatusBadRequest}
	}

	sortField := pb.SortField_NAME
	sortDir := pb.SortDir_ASC
	qSortField := q.Get("sort_field")
	qSortDir := q.Get("sort_dir")

	if qSortField != "" || qSortDir != "" {
		sortField = enums.StringToPBEnum(
			qSortField,
			pb.SortField_SortField_value,
			pb.SortField_UNDEFINED,
		)
		sortDir = enums.StringToPBEnum(
			qSortDir,
			pb.SortDir_SortDir_value,
			pb.SortDir_UNDEFINED,
		)
		if sortField == pb.SortField_UNDEFINED || sortDir == pb.SortDir_UNDEFINED {
			return &common.HandlerRes{
				Error:      errs.ERR_CHOOSE_VALID_OPTION,
				StatusCode: http.StatusBadRequest,
			}
		}
	}

	res, err := h.ProductService.GetProductsWithSorting(sortField, sortDir)
	if err != nil {
		return &common.HandlerRes{Error: err, StatusCode: http.StatusInternalServerError}
	}

	return &common.HandlerRes{
		Component: components.Base("Products", components.ProductTableView(res)),
	}
}
