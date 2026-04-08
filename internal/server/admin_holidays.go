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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminHolidaysListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays List Table Handler]"
	ctx := r.Context()

	holidays, err := s.services.holiday.GetAllHolidays(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "date, name, and type are required", http.StatusBadRequest)
		return
	}

	holidayType := enums.ParseHolidayTypeToEnum(holidayTypeStr)
	if holidayType == enums.HOLIDAY_TYPE_UNDEFINED {
		http.Error(w, "Invalid holiday type", http.StatusBadRequest)
		return
	}

	date, err := time.Parse(constants.DateLayoutISO, dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	_, err = s.services.holiday.CreateHoliday(ctx, date, name, holidayType)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Failed to create holiday", http.StatusInternalServerError)
		return
	}

	redirectHX(w, r, utils.URL("/admin/holidays"))
}

func (s *Server) adminHolidaysUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Update Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	name := r.FormValue("name")
	holidayTypeStr := r.FormValue("type")

	if idStr == "" || name == "" || holidayTypeStr == "" {
		http.Error(w, "id, name, and type are required", http.StatusBadRequest)
		return
	}

	id, err := time.Parse(constants.DateLayoutISO, idStr)
	if err != nil {
		http.Error(w, "Invalid id format", http.StatusBadRequest)
		return
	}

	holidayType := enums.ParseHolidayTypeToEnum(holidayTypeStr)
	if holidayType == enums.HOLIDAY_TYPE_UNDEFINED {
		http.Error(w, "Invalid holiday type", http.StatusBadRequest)
		return
	}

	existing, err := s.services.holiday.IsHoliday(ctx, id)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existing == nil {
		http.Error(w, "Holiday not found", http.StatusNotFound)
		return
	}

	_, err = s.services.holiday.UpdateHoliday(ctx, id.Unix(), name, holidayType)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Failed to update holiday", http.StatusInternalServerError)
		return
	}

	redirectHX(w, r, utils.URL("/admin/holidays"))
}

func (s *Server) adminHolidaysDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Delete Handler]"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		http.Error(w, "Invalid id format", http.StatusBadRequest)
		return
	}

	err := s.services.holiday.DeleteHoliday(ctx, id)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Failed to delete holiday", http.StatusInternalServerError)
		return
	}

	redirectHX(w, r, utils.URL("/admin/holidays"))
}
