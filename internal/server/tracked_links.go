package server

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) handleTrackedLink(w http.ResponseWriter, r *http.Request) {
	const logtag = "[HandleTrackedLink]"
	const page = "/"

	var pathReq forms.TrackedLinkPath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		redirectHX(w, r, utils.URL(page))
		return
	}
	slug := pathReq.Slug

	ctx := r.Context()
	link, err := s.services.trackedLink.GetTrackedLinkBySlug(ctx, slug)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URL(page))
		return
	}
	if link == nil {
		redirectHX(w, r, utils.URL(page))
		return
	}

	switch link.Status {
	case enums.TRACKED_LINK_STATUS_DRAFT, enums.TRACKED_LINK_STATUS_DELETED:
		redirectHX(w, r, utils.URL(page))
		return
	}

	var utmReq forms.TrackedLinkUTMQuery
	_ = httputil.BindQuery(r, &utmReq)

	go func() {
		ua := utils.ParseUserAgent(r.UserAgent())
		_ = s.services.trackedLink.RecordClick(
			context.Background(),
			slug,
			r.Referer(),
			r.UserAgent(),
			hashIP(r.RemoteAddr),
			ua.Device,
			utmReq.UTMSource,
			utmReq.UTMMedium,
			utmReq.UTMCampaign,
		)
	}()

	redirectHX(w, r, link.DestinationURL)
}

// TODO: Move to utils
func hashIP(remoteAddr string) string {
	hash := sha256.Sum256([]byte(remoteAddr))
	return hex.EncodeToString(hash[:])
}

func (s *Server) handleTrackedLinkQR(w http.ResponseWriter, r *http.Request) {
	const logtag = "[HandleTrackedLinkQR]"
	const page = "/admin/tracked-links"
	ctx := r.Context()

	var pathReq forms.TrackedLinkQRPath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrNotFound.Error()))
		return
	}
	if _, err := httputil.RequireEncodedID(s.encoder, pathReq.ID); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrNotFound.Error()))
		return
	}

	link, err := s.services.trackedLink.GetTrackedLinkByID(ctx, pathReq.ID)
	if err != nil || link == nil {
		err = cmp.Or(err, errs.ErrDBNil)
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	qrURL := fmt.Sprintf("%s/l/%s", utils.FullURL(""), link.Slug)

	sfKey := "qr:" + link.Slug
	res, err, _ := s.SF.Do(sfKey, func() (any, error) {
		return s.services.qr.GenerateQR(ctx, qrURL)
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qrBytes := res.([]byte)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.png", link.Slug))
	if _, err := w.Write(qrBytes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
