package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminCategoriesListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Categories List Page Handler]"
	const page = "/admin/categories"
	ctx := r.Context()

	if err := compadmin.AdminCategoriesListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminCategoriesListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Categories List Table Handler]"
	const page = "/admin/categories"
	ctx := r.Context()

	var q forms.AdminCategoriesListQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	listPage := httputil.PageOrDefault(q.Page, 1)

	categories, totalCount, listPage, err := s.services.productCategory.GetCategoriesListingPaginated(
		ctx,
		q.Search,
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

	var q forms.AdminCategoriesSubcategoriesQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	categoryName := q.Category

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

	var f forms.AdminCategoriesCreateForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	mode := f.Mode
	categoryName := f.CategoryName
	if mode == "existing" {
		categoryName = f.Category
	}

	subcategories := f.Subcategories

	if mode == "" || categoryName == "" || len(subcategories) == 0 {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrAllFieldsRequired.Error()))
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

	requests.InvalidateAdminCategoriesCache(s.cache)

	redirectHX(w, r, utils.URLWithSuccess(page, "Category created successfully"))
}
