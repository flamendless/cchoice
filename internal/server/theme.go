package server

import (
	"context"
	"net/http"

	compshop "cchoice/cmd/web/components/shop"
	"cchoice/cmd/web/models"
	"cchoice/internal/logs"
	"cchoice/internal/services"

	"go.uber.org/zap"
)

func (s *Server) themePreviewHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Theme Preview Handler]"
	ctx := r.Context()

	configuration := s.services.theme.PreviewConfigurationFromQuery(ctx, r.URL.Query(), logtag)

	homePageData := models.HomePageData{
		Sections: models.BuildPostHomeContentSections(s.GetBrandLogoCDNURL),
		ThemeCSS: services.BuildThemeCSS(configuration),
	}

	if err := compshop.HomePage(homePageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) activeThemeCSS(ctx context.Context, logtag string) string {
	theme, err := s.services.theme.GetActiveTheme(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("message", "failed to get active theme"))
		return ""
	}
	if theme == nil {
		return ""
	}

	configuration, err := services.UnmarshalThemeConfiguration(theme.Configuration, theme.ConfigurationType)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("message", "failed to parse active theme configuration"))
		return ""
	}
	logs.LogCtx(ctx).Debug(logtag, zap.String("theme_configuration", theme.Configuration), zap.Any("parsed_theme_configuration", configuration))
	return services.BuildThemeCSS(configuration)
}
