package server

import (
	"cmp"
	"net/http"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) adminTrackedLinksEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links Edit Page Handler]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	var p forms.AdminTrackedLinkPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	link, err := s.services.trackedLink.GetTrackedLinkByID(ctx, idStr)
	if err != nil || link == nil {
		err = cmp.Or(err, errs.ErrDBNil)
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	clickCount, _ := s.services.trackedLink.GetClickCount(ctx, link.ID)

	item := models.AdminTrackedLinkListItem{
		ID:             link.ID,
		Name:           link.Name,
		Slug:           link.Slug,
		DestinationURL: link.DestinationURL,
		Source:         link.Source,
		Medium:         link.Medium,
		Campaign:       link.Campaign.String,
		Clicks:         clickCount,
		Status:         link.Status,
	}

	if err := compadmin.AdminTrackedLinksEditModal(item).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminTrackedLinksCreateModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links Create Modal Handler]"
	const page = "/admin/tracked-links"

	if err := compadmin.AdminTrackedLinksCreateModal().Render(r.Context(), w); err != nil {
		logs.LogCtx(r.Context()).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminTrackedLinksListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links List Page Handler]"
	const page = "/admin/tracked-links"

	if err := compadmin.AdminTrackedLinksListPage("Tracked Links").Render(r.Context(), w); err != nil {
		logs.LogCtx(r.Context()).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminTrackedLinksListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links List Table Handler]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	links, err := s.services.trackedLink.ListTrackedLinks(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	linkList := make([]models.AdminTrackedLinkListItem, 0, len(links))
	for _, l := range links {
		clickCount, _ := s.services.trackedLink.GetClickCount(ctx, l.ID)
		linkList = append(linkList, models.AdminTrackedLinkListItem{
			ID:             l.ID,
			Name:           l.Name,
			Slug:           l.Slug,
			DestinationURL: l.DestinationURL,
			Source:         l.Source,
			Medium:         l.Medium,
			Campaign:       l.Campaign.String,
			Clicks:         clickCount,
			Status:         l.Status,
		})
	}

	if err := compadmin.AdminTrackedLinksListTable(linkList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminTrackedLinksCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links Create Handler]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	var f forms.AdminTrackedLinkForm
	if err := httputil.BindForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	name := f.Name
	slug := f.Slug
	destinationURL := f.DestinationURL
	sourceStr := f.Source
	mediumStr := f.Medium
	campaign := f.Campaign

	if name == "" || slug == "" || destinationURL == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMissingField.Error()))
		return
	}

	if !strings.HasPrefix(destinationURL, constants.PrefixHTTPS) {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidFormat.Error()))
		return
	}

	source := enums.ParseTrackedLinkSourceToEnum(sourceStr)
	medium := enums.ParseTrackedLinkMediumToEnum(mediumStr)

	if _, err := s.services.trackedLink.CreateTrackedLink(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		name,
		slug,
		destinationURL,
		source,
		medium,
		campaign,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Tracked link created successfully"))
}

func (s *Server) adminTrackedLinksUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links Update Handler]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	var p forms.AdminTrackedLinkPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	var f forms.AdminTrackedLinkForm
	if err := httputil.BindForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	name := f.Name
	slug := f.Slug
	destinationURL := f.DestinationURL
	sourceStr := f.Source
	mediumStr := f.Medium
	campaign := f.Campaign
	statusStr := f.Status

	if idStr == "" || name == "" || slug == "" || destinationURL == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMissingField.Error()))
		return
	}

	if !strings.HasPrefix(destinationURL, constants.PrefixHTTPS) {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidFormat.Error()))
		return
	}

	source := enums.ParseTrackedLinkSourceToEnum(sourceStr)
	medium := enums.ParseTrackedLinkMediumToEnum(mediumStr)
	status := enums.ParseTrackedLinkStatusToEnum(statusStr)

	if err := s.services.trackedLink.UpdateTrackedLink(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		idStr,
		name,
		slug,
		destinationURL,
		source,
		medium,
		campaign,
		status,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Tracked link updated successfully"))
}

func (s *Server) adminTrackedLinksDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links Delete Handler]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	var p forms.AdminTrackedLinkPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	if err := s.services.trackedLink.DeleteTrackedLink(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		idStr,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Tracked link deleted successfully"))
}
