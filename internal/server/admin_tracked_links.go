package server

import (
	"net/http"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminTrackedLinksEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links Edit Page Handler]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "Invalid ID"))
		return
	}

	link, err := s.services.trackedLink.GetTrackedLinkByID(ctx, idStr)
	if err != nil {
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

	if err := compadmin.AdminTrackedLinksCreateModal().Render(r.Context(), w); err != nil {
		logs.LogCtx(r.Context()).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/tracked-links", err.Error()))
		return
	}
}

func (s *Server) adminTrackedLinksListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links List Page Handler]"

	if err := compadmin.AdminTrackedLinksListPage("Tracked Links").Render(r.Context(), w); err != nil {
		logs.LogCtx(r.Context()).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/tracked-links", err.Error()))
		return
	}
}

func (s *Server) adminTrackedLinksListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links List Table Handler]"
	ctx := r.Context()

	links, err := s.services.trackedLink.ListTrackedLinks(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/tracked-links", err.Error()))
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
		redirectHX(w, r, utils.URLWithError("/admin/tracked-links", err.Error()))
		return
	}
}

func (s *Server) adminTrackedLinksCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Tracked Links Create Handler]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	name := r.FormValue("name")
	slug := r.FormValue("slug")
	destinationURL := r.FormValue("destination_url")
	sourceStr := r.FormValue("source")
	mediumStr := r.FormValue("medium")
	campaign := r.FormValue("campaign")

	if name == "" || slug == "" || destinationURL == "" {
		redirectHX(w, r, utils.URLWithError(page, "name, slug, and destination_url are required"))
		return
	}

	if !strings.HasPrefix(destinationURL, constants.PrefixHTTPS) {
		redirectHX(w, r, utils.URLWithError(page, "Destination URL must start with 'https://'"))
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

	idStr := chi.URLParam(r, "id")
	name := r.FormValue("name")
	slug := r.FormValue("slug")
	destinationURL := r.FormValue("destination_url")
	sourceStr := r.FormValue("source")
	mediumStr := r.FormValue("medium")
	campaign := r.FormValue("campaign")
	statusStr := r.FormValue("status")

	if idStr == "" || name == "" || slug == "" || destinationURL == "" {
		redirectHX(w, r, utils.URLWithError(page, "id, name, slug, and destination_url are required"))
		return
	}

	if !strings.HasPrefix(destinationURL, constants.PrefixHTTPS) {
		redirectHX(w, r, utils.URLWithError(page, "Destination URL must start with 'https://'"))
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

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "id is required"))
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
