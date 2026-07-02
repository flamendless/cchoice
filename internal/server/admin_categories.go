package server

import (
	"net/http"
	"strconv"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminCategoriesListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Categories List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminCategoriesListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/categories", err.Error()))
		return
	}
}

func (s *Server) adminCategoriesListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Categories List Table Handler]"
	const page = "/admin/categories"
	ctx := r.Context()

	search := strings.TrimSpace(r.URL.Query().Get("search"))

	listPage := 1
	if paramPage := r.URL.Query().Get("page"); paramPage != "" {
		if parsed, err := strconv.Atoi(paramPage); err == nil && parsed > 0 {
			listPage = parsed
		}
	}

	categories, totalCount, listPage, err := s.services.productCategory.GetCategoriesListingPaginated(
		ctx,
		search,
		listPage,
		constants.DefaultAdminTablePageSize,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/admin/categories/table"),
		Include:       "[name='search']",
		ContentTarget: "#categories-table-content",
	}

	if err := compadmin.AdminCategoriesTableContent(categories, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminCategoriesSubcategoriesHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Categories Subcategories Handler]"
	const page = "/admin/categories"
	ctx := r.Context()

	categoryName := strings.TrimSpace(r.URL.Query().Get("category"))
	if categoryName == "" {
		redirectHX(w, r, utils.URLWithError(page, "Category is required"))
		return
	}

	rows, err := s.services.productCategory.GetSubcategoriesForCategory(ctx, categoryName)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if err := compadmin.CategorySubcategoryRows(rows).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminCategoriesCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Categories Create Page Handler]"
	const page = "/admin/categories"
	ctx := r.Context()

	categoryNames, err := s.services.productCategory.GetAllCategoryNames(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if err := compadmin.CategoryCreateModal(categoryNames).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminCategoriesCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Categories Create Handler]"
	const page = "/admin/categories"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	mode := r.FormValue("mode")
	categoryName := strings.TrimSpace(r.FormValue("category_name"))
	if mode == "existing" {
		categoryName = strings.TrimSpace(r.FormValue("category"))
	}

	subcategories := r.Form["subcategories[]"]
	if len(subcategories) == 0 {
		subcategories = r.Form["subcategories"]
	}

	if mode == "" || categoryName == "" || len(subcategories) == 0 {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	if err := s.services.productCategory.CreateCategories(ctx, s.sessionManager.GetString(ctx, SessionStaffID), services.CreateCategoriesParams{
		Mode:          mode,
		CategoryName:  categoryName,
		Subcategories: subcategories,
	}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Category created successfully"))
}
