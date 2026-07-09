package server

import (
	"net/http"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const adminThemesPage = "/admin/themes"

func (s *Server) adminThemesListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Themes List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminThemesListPage("Themes").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, err.Error()))
		return
	}
}

func (s *Server) adminThemesCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Themes Create Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminThemesCreatePage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, err.Error()))
		return
	}
}

func (s *Server) adminThemesListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Themes List Table Handler]"
	ctx := r.Context()

	search := r.URL.Query().Get("search")
	sortBy, sortDir, err := utils.ParseListingSortQuery(r.URL.Query(), "TITLE", "START_DATE", "END_DATE", "STATUS")
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("sort_by", sortBy), zap.String("sort_dir", sortDir.String()), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, err.Error()))
		return
	}

	serviceThemes, err := s.services.theme.GetAllThemes(ctx, search, sortBy, sortDir.String())
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, err.Error()))
		return
	}

	themes := make([]models.AdminThemeListItem, 0, len(serviceThemes))
	for _, t := range serviceThemes {
		themes = append(themes, models.AdminThemeListItem{
			ID:                s.encoder.Encode(t.ID),
			Title:             t.Title,
			Status:            t.Status,
			StartDate:         t.StartDate,
			EndDate:           t.EndDate,
			ConfigurationType: t.ConfigurationType,
			Active:            t.Active,
			CreatedAt:         t.CreatedAt.Format(constants.DateTimeLayoutISO),
		})
	}

	if err := compadmin.AdminThemesListTable(themes).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, err.Error()))
		return
	}
}

func (s *Server) adminThemesCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Themes Create Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Failed to parse form"))
		return
	}

	title := r.FormValue("title")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	configTypeStr := r.FormValue("configuration_type")

	if title == "" || startDateStr == "" || endDateStr == "" || configTypeStr == "" {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "All fields are required"))
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid end date format"))
		return
	}

	configType := enums.ParseThemeConfigTypeToEnum(configTypeStr)
	if configType == enums.THEME_CONFIG_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid configuration type"))
		return
	}
	configuration := s.services.theme.ReadColorFields(r.Form)

	if _, err := s.services.theme.CreateTheme(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		title,
		startDate,
		endDate,
		configuration,
		configType,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminThemesPage, "Theme created successfully"))
}

func (s *Server) adminThemesEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Themes Edit Page Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid id format"))
		return
	}

	theme, err := s.services.theme.GetThemeByID(ctx, id)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Failed to get theme"))
		return
	}
	if theme == nil {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Theme not found"))
		return
	}
	if theme.Status == enums.THEME_STATUS_DELETED {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Deleted themes cannot be edited"))
		return
	}

	configuration, err := services.UnmarshalThemeConfiguration(theme.Configuration, theme.ConfigurationType)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Failed to parse theme configuration"))
		return
	}

	themeItem := models.AdminThemeListItem{
		ID:                s.encoder.Encode(theme.ID),
		Title:             theme.Title,
		Status:            theme.Status,
		StartDate:         theme.StartDate,
		EndDate:           theme.EndDate,
		Configuration:     configuration,
		ConfigurationType: theme.ConfigurationType,
		CreatedAt:         theme.CreatedAt.Format(constants.DateTimeLayoutISO),
	}

	if err := compadmin.ThemeEditModal(themeItem).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Failed to render edit form"))
		return
	}
}

func (s *Server) adminThemesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Themes Update Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Failed to parse form"))
		return
	}

	idStr := chi.URLParam(r, "id")
	title := r.FormValue("title")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	configTypeStr := r.FormValue("configuration_type")
	statusStr := r.FormValue("status")

	if idStr == "" || title == "" || startDateStr == "" || endDateStr == "" || configTypeStr == "" || statusStr == "" {
		logs.Log().Warn(
			logtag,
			zap.String("id", idStr),
			zap.Any("form value", r.Form),
		)
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "All fields are required"))
		return
	}

	themeStatus := enums.ParseThemeStatusToEnum(statusStr)
	switch themeStatus {
	case enums.THEME_STATUS_UNDEFINED:
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid theme status"))
		return
	case enums.THEME_STATUS_DELETED:
		s.adminThemesDeleteHandler(w, r)
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid end date format"))
		return
	}

	configType := enums.ParseThemeConfigTypeToEnum(configTypeStr)
	if configType == enums.THEME_CONFIG_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Invalid configuration type"))
		return
	}
	configuration := s.services.theme.ReadColorFields(r.Form)

	if err := s.services.theme.UpdateTheme(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		idStr,
		title,
		themeStatus,
		startDate,
		endDate,
		configuration,
		configType,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminThemesPage, "Theme updated successfully"))
}

func (s *Server) adminThemesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Themes Delete Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if err := s.services.theme.DeleteTheme(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("id", idStr), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminThemesPage, "Failed to delete theme"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminThemesPage, "Theme deleted successfully"))
}
