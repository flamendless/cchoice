package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	pb "cchoice/proto"
	"context"
	"net/http"

	"github.com/a-h/templ"
	"go.uber.org/zap"
)

type SettingsService interface {
	pb.SettingsServiceClient
}

type SettingsHandler struct {
	Logger          *zap.Logger
	SettingsService SettingsService
}

func NewSettingsHandler(
	logger *zap.Logger,
	settingsService SettingsService,
) SettingsHandler {
	return SettingsHandler{
		Logger:          logger,
		SettingsService: settingsService,
	}
}

func (h SettingsHandler) FooterDetails(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	res, err := h.SettingsService.GetSettingsByNames(context.TODO(), &pb.SettingsByNamesRequest{
		Names: []string{
			"url_tiktok",
			"url_facebook",
			"url_gmap",
			"email",
			"mobile_no",
		},
	})
	if err != nil {
		return &common.HandlerRes{Error: err}
	}
	return &common.HandlerRes{
		Component: components.Footer(common.FooterDetails{
			URLTikTok:   string(templ.URL(res.Settings["url_tiktok"])),
			URLFacebook: string(templ.URL(res.Settings["url_facebook"])),
			URLGMap:     string(templ.URL(res.Settings["url_gmap"])),
			Email:       string(templ.URL(res.Settings["email"])),
			MobileNo:    string(templ.URL(res.Settings["mobile_no"])),
		}),
	}
}
