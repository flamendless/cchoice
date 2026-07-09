package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/encode"
	"cchoice/internal/encode/b64"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminCPointsGeneratePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin C-Points Generate Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminCPointsGeneratePage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminCPointsGeneratePostHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin C-Points Generate Post Handler]"
	const page = "/admin/cpoints/generate"
	ctx := r.Context()

	var f forms.AdminCPointsGenerateForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	customerID, err := httputil.RequireEncodedID(s.encoder, f.CustomerID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}
	value := f.Value

	var expiresAtTime *time.Time
	if f.ExpiresAt != "" {
		now := time.Now()
		switch f.ExpiresAt {
		case "1_week":
			t := now.AddDate(0, 0, 7)
			expiresAtTime = &t
		case "1_month":
			t := now.AddDate(0, 1, 0)
			expiresAtTime = &t
		case "1_year":
			t := now.AddDate(1, 0, 0)
			expiresAtTime = &t
		}
	}

	var productSkus []string
	if f.ProductSkus != "" {
		parts := strings.SplitSeq(f.ProductSkus, ",")
		for part := range parts {
			sku := strings.TrimSpace(part)
			if sku != "" {
				productSkus = append(productSkus, sku)
			}
		}
	}

	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	customerIDDecoded := s.encoder.Decode(customerID)
	if customerIDDecoded == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	cpoint, err := s.services.cpoint.CreateCpoint(context.Background(), services.CreateCpointParams{
		StaffID:     staffIDStr,
		CustomerID:  customerID,
		Value:       value,
		ProductSkus: productSkus,
		ExpiresAt:   expiresAtTime,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrCpointGenerateFailed.Error()))
		return
	}

	redemptionURL := s.services.cpoint.GenerateRedemptionURL(cpoint.Code, customerIDDecoded)
	_, err = s.services.qr.GenerateQR(ctx, redemptionURL)
	if err != nil {
		logs.LogCtx(ctx).Warn("failed to generate QR code", zap.Error(err))
	}

	redirectURL := utils.URLWithParams("/admin/cpoints/code", map[string]string{
		"code":        cpoint.Code,
		"redemption":  redemptionURL,
		"customer_id": customerID,
	})
	redirectHX(w, r, redirectURL)
}

func (s *Server) adminCPointsCodePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin C-Points Code Page Handler]"
	const page = "/admin/cpoints/generate"
	ctx := r.Context()

	var q forms.AdminCPointsCodeQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		redirectHX(w, r, utils.URL(page))
		return
	}
	code := q.Code
	redemptionURL := q.Redemption

	if err := compadmin.AdminCPointsCodePage(code, redemptionURL).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminCPointsQRHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin C-Points QR Handler]"
	ctx := r.Context()

	var q forms.AdminCPointsQRQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}
	code := q.Code

	cpoint, err := s.services.cpoint.GetCpointByCode(ctx, code)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "cpoint not found", http.StatusNotFound)
		return
	}

	redemptionURL := s.services.cpoint.GenerateRedemptionURL(code, cpoint.CustomerID)
	qrBytes, err := s.services.qr.GenerateQR(ctx, redemptionURL)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "failed to generate QR", http.StatusInternalServerError)
		return
	}

	imgfmt := enums.IMAGE_FORMAT_PNG.DataURIPrefix()
	qrBase64 := imgfmt + b64.ToBase64(qrBytes)

	if err := compadmin.CPointsQRImage(qrBase64).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}
