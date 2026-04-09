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
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminHolidaysListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminHolidaysListPage("Holidays").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", err.Error()))
		return
	}
}

func (s *Server) adminHolidaysListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays List Table Handler]"
	ctx := r.Context()

	holidays, err := s.services.holiday.GetAllHolidays(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Internal server error"))
		return
	}

	holidayList := make([]models.AdminHolidayListItem, 0, len(holidays))
	for _, h := range holidays {
		holidayList = append(holidayList, models.AdminHolidayListItem{
			ID:   s.encoder.Encode(h.ID),
			Date: h.Date,
			Name: h.Name,
			Type: h.Type,
		})
	}

	if err := compadmin.AdminHolidaysListTable(holidayList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Internal server error"))
		return
	}
}

func (s *Server) adminHolidaysCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Create Handler]"
	ctx := r.Context()

	dateStr := r.FormValue("date")
	name := r.FormValue("name")
	holidayTypeStr := r.FormValue("type")

	if dateStr == "" || name == "" || holidayTypeStr == "" {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "date, name, and type are required"))
		return
	}

	holidayType := enums.ParseHolidayTypeToEnum(holidayTypeStr)
	if holidayType == enums.HOLIDAY_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Invalid holiday type"))
		return
	}

	date, err := time.Parse(constants.DateLayoutISO, dateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Invalid date format"))
		return
	}

	if _, err = s.services.holiday.CreateHoliday(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		date,
		name,
		holidayType,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Failed to create holiday"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess("/admin/holidays", "Holiday created successfully"))
}

func (s *Server) adminHolidaysUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Update Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	name := r.FormValue("name")
	holidayTypeStr := r.FormValue("type")

	if idStr == "" || name == "" || holidayTypeStr == "" {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "id, name, and type are required"))
		return
	}

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Invalid id format"))
		return
	}

	holidayType := enums.ParseHolidayTypeToEnum(holidayTypeStr)
	if holidayType == enums.HOLIDAY_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Invalid holiday type"))
		return
	}

	_, err := s.services.holiday.UpdateHoliday(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		id,
		name,
		holidayType,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Failed to update holiday"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess("/admin/holidays", "Holiday updated successfully"))
}

func (s *Server) adminHolidaysDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Delete Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Invalid id format"))
		return
	}

	if err := s.services.holiday.DeleteHoliday(ctx, s.sessionManager.GetString(ctx, SessionStaffID), id); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Failed to delete holiday"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess("/admin/holidays", "Holiday deleted successfully"))
}

func (s *Server) adminHolidaysEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Edit Page Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Invalid id format"))
		return
	}

	holidays, err := s.services.holiday.GetAllHolidays(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Internal server error"))
		return
	}

	var holidayItem *models.AdminHolidayListItem
	for _, h := range holidays {
		if h.ID == id {
			holidayItem = &models.AdminHolidayListItem{
				ID:   s.encoder.Encode(h.ID),
				Date: h.Date,
				Name: h.Name,
				Type: h.Type,
			}
			break
		}
	}

	if holidayItem == nil {
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Holiday not found"))
		return
	}

	w.Header().Set("HX-Reswap", "innerHTML")
	if err := compadmin.HolidayEditModal(*holidayItem).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/holidays", "Failed to render edit form"))
		return
	}
}
