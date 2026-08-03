package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminPromosListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos List Page Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	if err := compadmin.AdminPromosListPage("Promos").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminPromosCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Create Page Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	trackedLinks, err := s.services.trackedLink.ListTrackedLinks(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if err := compadmin.PromoCreateModal(toAdminTrackedLinkOptions(trackedLinks)).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
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
	trackedLinks, err := s.services.trackedLink.ListTrackedLinks(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	trackedLinkSlugs := trackedLinkSlugMap(trackedLinks)

	for _, p := range servicePromos {
		promos = append(promos, toAdminPromoListItem(s.encoder, p, trackedLinkSlugs))
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
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(errs.ErrInvalidParams)))
		return
	}

	var f forms.AdminPromoForm
	if err := httputil.BindMultipartForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrAllFieldsRequired.Error()))
		return
	}
	title := f.Title
	description := f.Description
	mediaURL := f.MediaURL
	startDateStr := f.StartDate
	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	endDateStr := f.EndDate
	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	if startDate.After(endDate) {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrValidationStartEndDates.Error()))
		return
	}

	promoTypeStr := f.Type
	promoType := enums.MustParsePromoTypeToEnum(promoTypeStr)
	if promoType == enums.PROMO_TYPE_BANNER_VIDEO && mediaURL == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrPromoMediaURLRequired.Error()))
		return
	}

	if title == "" || description == "" || startDateStr == "" || endDateStr == "" || promoTypeStr == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrAllFieldsRequired.Error()))
		return
	}

	bannerOnly := f.BannerOnly == "on"
	priority := f.Priority

	if promoType == enums.PROMO_TYPE_BANNER_IMAGE {
		file, header, err := r.FormFile("media_file")
		if err != nil {
			redirectHX(w, r, utils.URLWithError(page, errs.ErrPromoMediaFileRequired.Error()))
			return
		}
		defer file.Close()

		filename := s.services.image.GenerateFilename(
			enums.IMAGE_PREFIX_PROMO_IMAGE,
			filepath.Ext(header.Filename),
			title,
		)
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, file); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, errs.ErrFileRead.Error()))
			return
		}

		contentType := header.Header.Get("Content-Type")
		url, err := s.services.image.UploadPromoBannerImage(ctx, title, filename, &buf, contentType)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
		mediaURL = url
	}

	linkFields, err := services.ParsePromoLinkFields(
		ctx,
		promoType,
		enums.ParsePromoLinkTypeToEnum(f.LinkType),
		f.TrackedLinkID,
		f.LinkURL,
		s.services.trackedLink,
	)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
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
		bannerOnly,
		priority,
		linkFields,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Promo created successfully"))
}

func (s *Server) adminPromosEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Edit Page Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	var p forms.AdminPromoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	promo, err := s.services.promo.GetPromoByID(ctx, id)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrPromoGetFailed.Error()))
		return
	}

	if promo == nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrPromoNotFound.Error()))
		return
	}

	trackedLinks, err := s.services.trackedLink.ListTrackedLinks(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	promoItem := toAdminPromoListItem(s.encoder, *promo, trackedLinkSlugMap(trackedLinks))

	if err := compadmin.PromoEditModal(promoItem, toAdminTrackedLinkOptions(trackedLinks)).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrRenderFailed.Error()))
		return
	}
}

func (s *Server) adminPromosUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Promos Update Handler]"
	const page = "/admin/promos"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(errs.ErrInvalidParams)))
		return
	}

	var p forms.AdminPromoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrAllFieldsRequired.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrAllFieldsRequired.Error()))
		return
	}
	var f forms.AdminPromoForm
	if err := httputil.BindMultipartForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrAllFieldsRequired.Error()))
		return
	}
	title := f.Title
	description := f.Description
	mediaURL := f.MediaURL
	startDateStr := f.StartDate
	endDateStr := f.EndDate
	promoTypeStr := f.Type
	promoStatusStr := f.Status

	if idStr == "" || title == "" || description == "" || startDateStr == "" || endDateStr == "" || promoTypeStr == "" || promoStatusStr == "" {
		logs.Log().Warn(
			logtag,
			zap.String("id", idStr),
			zap.Any("form value", r.Form),
		)
		redirectHX(w, r, utils.URLWithError(page, errs.ErrAllFieldsRequired.Error()))
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	if startDate.After(endDate) {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrValidationStartEndDates.Error()))
		return
	}

	promoType := enums.MustParsePromoTypeToEnum(promoTypeStr)
	promoStatus := enums.MustParsePromoStatusToEnum(promoStatusStr)

	if promoStatus == enums.PROMO_STATUS_DELETED {
		s.adminPromosDeleteHandler(w, r)
		return
	}

	bannerOnly := f.BannerOnly == "on"
	priority := f.Priority

	if promoType == enums.PROMO_TYPE_BANNER_IMAGE {
		file, header, err := r.FormFile("media_file")
		if err == nil {
			defer file.Close()

			filename := s.services.image.GenerateFilename(
				enums.IMAGE_PREFIX_PROMO_IMAGE,
				filepath.Ext(header.Filename),
				title,
			)
			buf := bytes.Buffer{}
			if _, err := io.Copy(&buf, file); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, errs.ErrFileRead.Error()))
				return
			}

			contentType := header.Header.Get("Content-Type")
			url, err := s.services.image.UploadPromoBannerImage(ctx, title, filename, &buf, contentType)
			if err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, err.Error()))
				return
			}
			mediaURL = url
		}
	}

	linkFields, err := services.ParsePromoLinkFields(
		ctx,
		promoType,
		enums.ParsePromoLinkTypeToEnum(f.LinkType),
		f.TrackedLinkID,
		f.LinkURL,
		s.services.trackedLink,
	)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

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
		bannerOnly,
		priority,
		linkFields,
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

	var p forms.AdminPromoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrPromoDeleteFailed.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrPromoDeleteFailed.Error()))
		return
	}
	if err := s.services.promo.DeletePromo(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrPromoDeleteFailed.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Promo deleted successfully"))
}

func trackedLinkSlugMap(links []services.TrackedLink) map[string]string {
	result := make(map[string]string, len(links))
	for _, l := range links {
		result[l.ID] = l.Slug
	}
	return result
}

func toAdminTrackedLinkOptions(links []services.TrackedLink) []models.AdminTrackedLinkOption {
	result := make([]models.AdminTrackedLinkOption, 0, len(links))
	for _, l := range links {
		result = append(result, models.AdminTrackedLinkOption{
			ID:     l.ID,
			Name:   l.Name,
			Slug:   l.Slug,
			Status: l.Status,
		})
	}
	return result
}

func toAdminPromoListItem(encoder encode.IEncode, p services.Promo, trackedLinkSlugs map[string]string) models.AdminPromoListItem {
	linkFields := services.PromoLinkFields{
		TrackedLinkID: p.TrackedLinkID,
		LinkURL:       p.LinkURL,
	}
	linkType := services.PromoLinkTypeFromFields(linkFields)
	linkLabel := promoLinkLabel(p.Type, linkFields, trackedLinkSlugs)

	return models.AdminPromoListItem{
		ID:            encoder.Encode(p.ID),
		Title:         p.Title,
		Description:   p.Description,
		MediaURL:      p.MediaURL,
		StartDate:     p.StartDate,
		EndDate:       p.EndDate,
		Type:          p.Type,
		Status:        p.Status,
		BannerOnly:    p.BannerOnly.Bool,
		Priority:      p.Priority.Int64,
		CreatedAt:     utils.FormatTimePH(p.CreatedAt),
		LinkType:      linkType,
		TrackedLinkID: p.TrackedLinkID,
		LinkURL:       p.LinkURL,
		LinkLabel:     linkLabel,
	}
}

func promoLinkLabel(promoType enums.PromoType, fields services.PromoLinkFields, trackedLinkSlugs map[string]string) string {
	if promoType == enums.PROMO_TYPE_BANNER_VIDEO {
		return "—"
	}
	if fields.TrackedLinkID != "" {
		if slug, ok := trackedLinkSlugs[fields.TrackedLinkID]; ok {
			return "/l/" + slug
		}
		return fields.TrackedLinkID
	}
	if fields.LinkURL != "" {
		return fields.LinkURL
	}
	return "—"
}

func resolvePromoHref(ctx context.Context, trackedLink *services.TrackedLinkService, p services.Promo) string {
	if p.Type != enums.PROMO_TYPE_BANNER_IMAGE {
		return ""
	}
	if p.TrackedLinkID != "" {
		link, err := trackedLink.GetTrackedLinkByID(ctx, p.TrackedLinkID)
		if err != nil || link == nil || link.Status == enums.TRACKED_LINK_STATUS_DELETED {
			return ""
		}
		return utils.URLf("/l/%s", link.Slug)
	}
	if p.LinkURL != "" {
		return utils.ResolveHref(p.LinkURL)
	}
	return ""
}
