package server

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminPromosListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminPromosListPage("Promos").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/promos", err.Error()))
		return
	}
}

func (s *Server) adminPromosCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Create Page Handler]"
	ctx := r.Context()

	if err := compadmin.PromoCreateModal().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/promos", err.Error()))
		return
	}
}

func (s *Server) adminPromosListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos List Table Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	servicePromos, err := s.services.promo.GetAllPromos(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	promos := make([]models.AdminPromoListItem, 0, len(servicePromos))
	for _, p := range servicePromos {
		promos = append(promos, models.AdminPromoListItem{
			ID:        s.encoder.Encode(p.ID),
			Title:     p.Title,
			MediaURL:  p.MediaURL,
			StartDate: p.StartDate,
			EndDate:   p.EndDate,
			Type:      p.Type,
			Status:    p.Status,
			CreatedAt: p.CreatedAt.Format(constants.DateTimeLayoutISO),
		})
	}

	if err := compadmin.AdminPromosListTable(promos).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminPromosCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Create Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	mediaURL := r.FormValue("media_url")

	startDateStr := r.FormValue("start_date")
	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid start date format"))
		return
	}

	endDateStr := r.FormValue("end_date")
	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid end date format"))
		return
	}

	promoTypeStr := r.FormValue("type")
	promoType := enums.MustParsePromoTypeToEnum(promoTypeStr)
	if promoType == enums.PROMO_TYPE_BANNER_VIDEO && mediaURL == "" {
		redirectHX(w, r, utils.URLWithError(page, "Media URL is required for video type"))
		return
	}

	if title == "" || description == "" || startDateStr == "" || endDateStr == "" || promoTypeStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	if promoType == enums.PROMO_TYPE_BANNER_IMAGE {
		file, header, err := r.FormFile("media_file")
		if err != nil {
			redirectHX(w, r, utils.URLWithError(page, "Media file is required for image type"))
			return
		}
		defer file.Close()

		filename := s.services.image.GenerateFilename(filepath.Ext(header.Filename), title)
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, file); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, "Failed to read file"))
			return
		}

		contentType := header.Header.Get("Content-Type")
		url, err := s.services.image.UploadBrandImage(ctx, title, filename, &buf, contentType)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
		mediaURL = url
	}

	if _, err := s.services.promo.CreatePromo(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		title,
		description,
		mediaURL,
		startDate,
		endDate,
		promoType,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Promo created successfully"))
}

func (s *Server) adminPromosEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Edit Page Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError("/admin/promos", "Invalid id format"))
		return
	}

	promo, err := s.services.promo.GetPromoByID(ctx, id)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/promos", "Failed to get promo"))
		return
	}

	if promo == nil {
		redirectHX(w, r, utils.URLWithError("/admin/promos", "Promo not found"))
		return
	}

	promoItem := models.AdminPromoListItem{
		ID:          s.encoder.Encode(promo.ID),
		Title:       promo.Title,
		Description: promo.Description,
		MediaURL:    promo.MediaURL,
		StartDate:   promo.StartDate,
		EndDate:     promo.EndDate,
		Type:        promo.Type,
		Status:      promo.Status,
		CreatedAt:   promo.CreatedAt.Format(constants.DateTimeLayoutISO),
	}

	if err := compadmin.PromoEditModal(promoItem).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/promos", "Failed to render edit form"))
		return
	}
}

func (s *Server) adminPromosUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Update Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	idStr := chi.URLParam(r, "id")
	title := r.FormValue("title")
	description := r.FormValue("description")
	mediaURL := r.FormValue("media_url")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	promoTypeStr := r.FormValue("type")
	promoStatusStr := r.FormValue("status")

	if idStr == "" || title == "" || description == "" || startDateStr == "" || endDateStr == "" || promoTypeStr == "" || promoStatusStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid end date format"))
		return
	}

	promoType := enums.MustParsePromoTypeToEnum(promoTypeStr)
	promoStatus := enums.MustParsePromoStatusToEnum(promoStatusStr)

	if err := s.services.promo.UpdatePromo(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		idStr,
		title,
		description,
		mediaURL,
		startDate,
		endDate,
		promoType,
		promoStatus,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Promo updated successfully"))
}

func (s *Server) adminPromosDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Delete Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if err := s.services.promo.DeletePromo(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to delete promo"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Promo deleted successfully"))
}
