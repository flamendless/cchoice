package server

import (
	"net/http"
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
	"cchoice/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) adminHolidaysListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays List Page Handler]"
	const page = "/admin/holidays"
	ctx := r.Context()

	if err := compadmin.AdminHolidaysListPage("Holidays").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminHolidaysListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays List Table Handler]"
	const page = "/admin/holidays"
	ctx := r.Context()

	holidays, err := s.services.holiday.GetAllHolidays(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInternalServer.Error()))
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
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInternalServer.Error()))
		return
	}
}

func (s *Server) adminHolidaysCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Create Handler]"
	const page = "/admin/holidays"
	ctx := r.Context()

	var f forms.AdminHolidayCreateForm
	if err := httputil.BindForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	dateStr := f.Date
	name := f.Name
	holidayTypeStr := f.Type

	holidayType := enums.ParseHolidayTypeToEnum(holidayTypeStr)
	if holidayType == enums.HOLIDAY_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrHolidayInvalidType.Error()))
		return
	}

	date, err := time.Parse(constants.DateLayoutISO, dateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
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
		redirectHX(w, r, utils.URLWithError(page, errs.ErrHolidayCreateFailed.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Holiday created successfully"))
}

func (s *Server) adminHolidaysUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Update Handler]"
	const page = "/admin/holidays"
	ctx := r.Context()

	var p forms.AdminHolidayPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}
	var f forms.AdminHolidayUpdateForm
	if err := httputil.BindForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	name := f.Name
	holidayTypeStr := f.Type

	id := s.encoder.Decode(idStr)
	if id == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	holidayType := enums.ParseHolidayTypeToEnum(holidayTypeStr)
	if holidayType == enums.HOLIDAY_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrHolidayInvalidType.Error()))
		return
	}

	_, err = s.services.holiday.UpdateHoliday(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		id,
		name,
		holidayType,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrHolidayUpdateFailed.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Holiday updated successfully"))
}

func (s *Server) adminHolidaysDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Delete Handler]"
	const page = "/admin/holidays"
	ctx := r.Context()

	var p forms.AdminHolidayPath
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

	if err := s.services.holiday.DeleteHoliday(ctx, s.sessionManager.GetString(ctx, SessionStaffID), id); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrHolidayDeleteFailed.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Holiday deleted successfully"))
}

func (s *Server) adminHolidaysEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Holidays Edit Page Handler]"
	const page = "/admin/holidays"
	ctx := r.Context()

	var p forms.AdminHolidayPath
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

	holidays, err := s.services.holiday.GetAllHolidays(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInternalServer.Error()))
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
		redirectHX(w, r, utils.URLWithError(page, errs.ErrHolidayNotFound.Error()))
		return
	}

	if err := compadmin.HolidayEditModal(*holidayItem).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrRenderFailed.Error()))
		return
	}
}
