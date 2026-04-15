package server

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminBrandsListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Brands List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminBrandsListPage("Brands").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/brands", err.Error()))
		return
	}
}

func (s *Server) adminBrandsCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Brands Create Page Handler]"
	ctx := r.Context()

	w.Header().Set("HX-Reswap", "innerHTML")
	if err := compadmin.BrandCreateModal().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/brands", err.Error()))
		return
	}
}

func (s *Server) adminBrandsListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Brands List Table Handler]"
	const page = "/admin/brands"
	ctx := r.Context()

	var brands []models.AdminBrandListItem

	searchQuery := r.URL.Query().Get("search")
	if searchQuery != "" {
		serviceBrands, err := s.services.brand.SearchBrandsByName(ctx, searchQuery)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}

		brands = make([]models.AdminBrandListItem, 0, len(serviceBrands))
		for _, b := range serviceBrands {
			brands = append(brands, models.AdminBrandListItem{
				ID:           s.encoder.Encode(b.ID),
				Name:         b.Name,
				LogoS3URL:    b.LogoS3URL,
				BrandImageID: s.encoder.Encode(b.BrandImageID),
				ProductCount: b.ProductCount,
				CreatedAt:    b.CreatedAt.Format(constants.DateTimeLayoutISO),
			})
		}
	} else {
		serviceBrands, err := s.services.brand.GetAllBrands(ctx)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
		brands = make([]models.AdminBrandListItem, 0, len(serviceBrands))
		for _, b := range serviceBrands {
			brands = append(brands, models.AdminBrandListItem{
				ID:           s.encoder.Encode(b.ID),
				Name:         b.Name,
				LogoS3URL:    b.LogoS3URL,
				BrandImageID: s.encoder.Encode(b.BrandImageID),
				ProductCount: b.ProductCount,
				CreatedAt:    b.CreatedAt.Format(constants.DateTimeLayoutISO),
			})
		}
	}

	if err := compadmin.AdminBrandsListTable(brands, searchQuery).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminBrandsCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Brands Create Handler]"
	const page = "/admin/brands"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	brandName := r.FormValue("name")
	if brandName == "" {
		redirectHX(w, r, utils.URLWithError(page, "Brand name is required"))
		return
	}

	var logoS3URL string
	if conf.Conf().Test.LocalUploadImage || conf.Conf().IsProd() {
		file, header, err := r.FormFile("logo")
		if err != nil {
			redirectHX(w, r, utils.URLWithError(page, "Logo image is required"))
			return
		}
		defer file.Close()

		filename := s.services.image.GenerateFilename(filepath.Ext(header.Filename), brandName)
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, file); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, "Failed to read image"))
			return
		}

		contentType := header.Header.Get("Content-Type")
		url, err := s.services.image.UploadBrandImage(
			ctx,
			brandName,
			filename,
			&buf,
			contentType,
		)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
		logoS3URL = url
	}

	if _, err := s.services.brand.CreateBrand(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		brandName,
		logoS3URL,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Brand created successfully"))
}

func (s *Server) adminBrandsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Brands Update Handler]"
	const page = "/admin/brands"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	brandID := chi.URLParam(r, "id")
	brandName := r.FormValue("name")

	if brandID == "" || brandName == "" {
		redirectHX(w, r, utils.URLWithError(page, "id and name are required"))
		return
	}

	if id := s.encoder.Decode(brandID); id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrDecode.Error()))
		return
	}

	var logoS3URL string
	if conf.Conf().Test.LocalUploadImage || conf.Conf().IsProd() {
		file, header, err := r.FormFile("logo")
		if err == nil {
			defer file.Close()

			filename := s.services.image.GenerateFilename(filepath.Ext(header.Filename), brandName)
			buf := bytes.Buffer{}
			if _, err := io.Copy(&buf, file); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, "Failed to read image"))
				return
			}

			contentType := header.Header.Get("Content-Type")
			url, err := s.services.image.UploadBrandImage(
				ctx,
				brandName,
				filename,
				&buf,
				contentType,
			)
			if err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, err.Error()))
				return
			}
			logoS3URL = url
		}
	}

	if err := s.services.brand.UpdateBrand(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		brandID,
		brandName,
		logoS3URL,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to update brand"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Brand updated successfully"))
}

func (s *Server) adminBrandsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Brands Delete Handler]"
	const page = "/admin/brands"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if err := s.services.brand.DeleteBrand(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to delete brand"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Brand deleted successfully"))
}

func (s *Server) adminBrandsEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Brands Edit Page Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError("/admin/brands", "Invalid id format"))
		return
	}

	serviceBrands, err := s.services.brand.GetAllBrands(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/brands", "Internal server error"))
		return
	}

	var brandItem *models.AdminBrandListItem
	for _, b := range serviceBrands {
		if b.ID == id {
			brandItem = &models.AdminBrandListItem{
				ID:           s.encoder.Encode(b.ID),
				Name:         b.Name,
				LogoS3URL:    b.LogoS3URL,
				BrandImageID: s.encoder.Encode(b.BrandImageID),
				ProductCount: b.ProductCount,
				CreatedAt:    b.CreatedAt.Format("2006-01-02"),
			}
			break
		}
	}

	if brandItem == nil {
		redirectHX(w, r, utils.URLWithError("/admin/brands", "Brand not found"))
		return
	}

	w.Header().Set("HX-Reswap", "innerHTML")
	if err := compadmin.BrandEditModal(*brandItem).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/brands", "Failed to render edit form"))
		return
	}
}
