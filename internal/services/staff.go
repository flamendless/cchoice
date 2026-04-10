package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type StaffService struct {
	encoder    encode.IEncode
	dbRO       database.IService
	dbRW       database.IService
	attendance *AttendanceService
	location   *LocationService
}

func NewStaffService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
) *StaffService {
	return &StaffService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
	}
}

func NewStaffServiceWithDeps(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	attendance *AttendanceService,
	location *LocationService,
) *StaffService {
	if attendance == nil || location == nil {
		panic("attendance and location services are required")
	}
	return &StaffService{
		encoder:    encoder,
		dbRO:       dbRO,
		dbRW:       dbRW,
		attendance: attendance,
		location:   location,
	}
}

func (s *StaffService) GetByID(ctx context.Context, staffID string) (models.AdminStaffProfile, error) {
	decodedID := s.encoder.Decode(staffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedID)
	if err != nil {
		return models.AdminStaffProfile{}, err
	}
	return s.BuildProfile(staff), nil
}

func (s *StaffService) UpdatePassword(ctx context.Context, staffID string, password string) error {
	decodedID := s.encoder.Decode(staffID)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.dbRW.GetQueries().UpdateStaffPassword(ctx, queries.UpdateStaffPasswordParams{
		Password: string(hash),
		ID:       decodedID,
	})
	return err
}

func (s *StaffService) UpdateProfile(ctx context.Context, params UpdateProfileParams) error {
	if !strings.HasPrefix(params.MobileNo, constants.PHMobilePrefix) {
		return errs.ErrValidationInvalidMobileNumber
	}

	decodedID := s.encoder.Decode(params.ID)
	middleNameNull := sql.NullString{String: params.MiddleName, Valid: params.MiddleName != ""}

	_, err := s.dbRW.GetQueries().UpdateStaffProfile(ctx, queries.UpdateStaffProfileParams{
		FirstName:  params.FirstName,
		MiddleName: middleNameNull,
		LastName:   params.LastName,
		MobileNo:   params.MobileNo,
		Birthdate:  params.Birthdate,
		DateHired:  params.DateHired,
		ID:         decodedID,
	})
	return err
}

func (s *StaffService) GetAll(ctx context.Context, limit int64) ([]models.Staff, error) {
	staffs, err := s.dbRO.GetQueries().GetAllStaffs(ctx, limit)
	if err != nil {
		return nil, err
	}

	list := make([]models.Staff, 0, len(staffs))
	for _, staff := range staffs {
		list = append(list, models.Staff{
			ID:       s.encoder.Encode(staff.ID),
			FullName: utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		})
	}
	return list, nil
}

func (s *StaffService) GetAllForAdmin(ctx context.Context, search string) ([]queries.GetAllStaffsForAdminRow, error) {
	return s.dbRO.GetQueries().GetAllStaffsForAdmin(ctx, search)
}

func (s *StaffService) GetCurrentStaff(ctx context.Context, staffID string) (models.AdminStaffProfile, error) {
	decodedID := s.encoder.Decode(staffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedID)
	if err != nil {
		return models.AdminStaffProfile{}, err
	}
	return s.BuildProfile(staff), nil
}

func (s *StaffService) BuildProfile(staff queries.GetStaffByIDRow) models.AdminStaffProfile {
	return models.AdminStaffProfile{
		ID:               s.encoder.Encode(staff.ID),
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		FirstName:        staff.FirstName,
		MiddleName:       staff.MiddleName.String,
		LastName:         staff.LastName,
		Birthdate:        staff.Birthdate,
		DateHired:        staff.DateHired,
		Position:         staff.Position,
		Email:            staff.Email,
		MobileNo:         staff.MobileNo,
		ScheduledTimeIn:  staff.TimeInSchedule.String,
		ScheduledTimeOut: staff.TimeOutSchedule.String,
		RequireInShop:    staff.RequireInShop,
		UserType:         enums.ParseStaffUserTypeToEnum(staff.UserType),
	}
}

func (s *StaffService) GetCurrentStaffWithAttendance(
	ctx context.Context,
	staffID string,
	sessionManager *scs.SessionManager,
) (models.AdminStaffProfile, error) {
	profile, err := s.GetCurrentStaff(ctx, staffID)
	if err != nil {
		return profile, err
	}

	today := time.Now().Format(constants.DateLayoutISO)
	profile.SelectedDate = today
	profile.CurrentDate = time.Now().Format(constants.DateLayoutDisplay)
	profile.CurrentTime = time.Now().Format(constants.TimeLayoutDisplay)

	if s.attendance == nil || s.location == nil {
		logs.LogCtx(ctx).Error("[StaffService] attendance or location service not initialized")
		return profile, nil
	}

	profile.CanTimeIn = true
	profile.CanTimeOut = false
	profile.CanLunchBreakIn = false
	profile.CanLunchBreakOut = false

	dayAtt, err := s.attendance.GetStaffDayAttendance(ctx, staffID, today)
	if err != nil {
		if err != sql.ErrNoRows {
			logs.LogCtx(ctx).Error("[StaffService] GetStaffDayAttendance", zap.Error(err))
		}
		return profile, nil
	}

	hasTimeIn := dayAtt.HasTimeIn
	hasTimeOut := dayAtt.HasTimeOut
	hasLunchBreakIn := dayAtt.HasLunchBreakIn
	hasLunchBreakOut := dayAtt.HasLunchBreakOut

	inShop, outShop := s.location.CheckShopRadius(ctx, sessionManager, dayAtt.InLocation, dayAtt.OutLocation)

	canTimeIn := !hasTimeIn
	canTimeOut := hasTimeIn && !hasTimeOut
	canLunchBreakIn := hasTimeIn && !hasLunchBreakIn
	canLunchBreakOut := !hasTimeOut && hasLunchBreakIn && !hasLunchBreakOut

	locationDisplay, distanceMeters := s.location.ComputeLocationDisplay(ctx, sessionManager)

	profile.HasTimeIn = hasTimeIn
	profile.HasTimeOut = hasTimeOut
	profile.CanTimeIn = canTimeIn
	profile.CanTimeOut = canTimeOut
	profile.CanLunchBreakIn = canLunchBreakIn
	profile.CanLunchBreakOut = canLunchBreakOut
	profile.MyAttendance = dayAtt.Computed
	profile.InShop = inShop
	profile.OutShop = outShop
	profile.LocationDisplay = locationDisplay
	profile.DistanceMeters = distanceMeters

	return profile, nil
}

func (s *StaffService) GetTimeOffs(ctx context.Context, staffID string) ([]models.StaffTimeOff, error) {
	decodedID := s.encoder.Decode(staffID)
	timeOffs, err := s.dbRO.GetQueries().GetStaffTimeOffsByStaffID(ctx, decodedID)
	if err != nil {
		return nil, err
	}

	staffTimeOffs := make([]models.StaffTimeOff, 0, len(timeOffs))
	for _, to := range timeOffs {
		var approvedBy string
		var approvedAt string

		if to.ApprovedBy.Valid && to.ApproverFirstName.Valid {
			approvedBy = utils.BuildFullName(
				to.ApproverFirstName.String,
				to.ApproverMiddleName.String,
				to.ApproverLastName.String,
			)
		} else {
			approvedBy = "-"
		}

		if to.ApprovedAt.Valid {
			approvedAt = to.ApprovedAt.Time.Format(constants.DateTimeLayoutISO)
		} else {
			approvedAt = "-"
		}

		staffTimeOffs = append(staffTimeOffs, models.StaffTimeOff{
			ID:          s.encoder.Encode(to.ID),
			Type:        enums.ParseTimeOffToEnum(to.Type),
			CreatedAt:   utils.ConvertToPH(to.CreatedAt),
			StartDate:   to.StartDate.Format(constants.DateLayoutISO),
			EndDate:     to.EndDate.Format(constants.DateLayoutISO),
			Description: to.Description,
			Approved:    to.Approved.Bool,
			ApprovedBy:  approvedBy,
			ApprovedAt:  approvedAt,
		})
	}
	return staffTimeOffs, nil
}

func (s *StaffService) Log() {
	logs.Log().Info("[StaffService] Loaded")
}

var _ IService = (*StaffService)(nil)
