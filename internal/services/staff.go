package services

import (
	"context"
	"database/sql"
	"errors"
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

	"golang.org/x/crypto/bcrypt"
)

type StaffService struct {
	encoder encode.IEncode
	dbRO    database.IService
	dbRW    database.IService
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

func (s *StaffService) GetByID(ctx context.Context, staffID string) (models.AdminStaffProfile, error) {
	decodedID := s.encoder.Decode(staffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedID)
	if err != nil {
		return models.AdminStaffProfile{}, err
	}

	return models.AdminStaffProfile{
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Birthdate:        staff.Birthdate,
		DateHired:        staff.DateHired,
		Position:         staff.Position,
		Email:            staff.Email,
		MobileNo:         staff.MobileNo,
		ScheduledTimeIn:  staff.TimeInSchedule.String,
		ScheduledTimeOut: staff.TimeOutSchedule.String,
		RequireInShop:    staff.RequireInShop,
		UserType:         enums.ParseStaffUserTypeToEnum(staff.UserType),
	}, nil
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

func (s *StaffService) Create(ctx context.Context, params CreateStaffParams) (string, error) {
	if params.FirstName == "" || params.LastName == "" || params.Position == "" ||
		params.Birthdate == "" || params.DateHired == "" || params.Email == "" ||
		params.MobileNo == "" || params.TimeInSchedule == "" || params.TimeOutSchedule == "" ||
		params.Password == "" {
		return "", errs.ErrMissingField
	}

	if _, err := time.Parse(constants.DateLayoutISO, params.Birthdate); err != nil {
		return "", errs.ErrInvalidFormat
	}

	if _, err := time.Parse(constants.DateLayoutISO, params.DateHired); err != nil {
		return "", errs.ErrInvalidFormat
	}

	sex := strings.ToUpper(params.Sex)
	if sex != "M" && sex != "F" {
		return "", errs.ErrInvalidInput
	}

	if params.UserType != enums.STAFF_USER_TYPE_STAFF && params.UserType != enums.STAFF_USER_TYPE_SUPERUSER {
		return "", errs.ErrInvalidInput
	}

	if !constants.ReEmail.MatchString(params.Email) {
		return "", errs.ErrInvalidFormat
	}

	mobileNo := params.MobileNo
	if !strings.HasPrefix(mobileNo, constants.PHMobilePrefix) {
		mobileNo = constants.PHMobilePrefix + mobileNo
	}
	if !constants.ReMobileNumber.MatchString(mobileNo) {
		return "", errs.ErrValidationInvalidMobileNumber
	}

	if _, err := time.Parse(constants.TimeLayoutHHMM, params.TimeInSchedule); err != nil {
		return "", errs.ErrInvalidFormat
	}

	if _, err := time.Parse(constants.TimeLayoutHHMM, params.TimeOutSchedule); err != nil {
		return "", errs.ErrInvalidFormat
	}

	if !constants.RePassword.MatchString(params.Password) {
		return "", errs.ErrInvalidFormat
	}

	if _, err := s.dbRO.GetQueries().GetStaffByEmail(ctx, params.Email); err == nil {
		return "", errs.ErrDuplicateEmail
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	middleNameNull := sql.NullString{String: params.MiddleName, Valid: params.MiddleName != ""}
	id, err := s.dbRW.GetQueries().CreateStaff(ctx, queries.CreateStaffParams{
		FirstName:       params.FirstName,
		MiddleName:      middleNameNull,
		LastName:        params.LastName,
		Birthdate:       params.Birthdate,
		Sex:             sex,
		DateHired:       params.DateHired,
		TimeInSchedule:  sql.NullString{Valid: true, String: params.TimeInSchedule},
		TimeOutSchedule: sql.NullString{Valid: true, String: params.TimeOutSchedule},
		Position:        params.Position,
		UserType:        params.UserType.String(),
		Email:           params.Email,
		MobileNo:        mobileNo,
		Password:        string(hash),
		RequireInShop:   params.RequireInShop,
	})
	if err != nil {
		return "", err
	}

	return s.encoder.Encode(id), nil
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
			Email:    staff.Email,
			Position: staff.Position,
			UserType: enums.ParseStaffUserTypeToEnum(staff.UserType),
		})
	}
	return list, nil
}

func (s *StaffService) GetAllForMemo(ctx context.Context, limit int64) ([]models.Staff, error) {
	staffs, err := s.dbRO.GetQueries().GetAllStaffsForMemo(ctx, limit)
	if err != nil {
		return nil, err
	}

	list := make([]models.Staff, 0, len(staffs))
	for _, staff := range staffs {
		list = append(list, models.Staff{
			ID:       s.encoder.Encode(staff.ID),
			FullName: utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
			Email:    staff.Email,
			Position: staff.Position,
			UserType: enums.ParseStaffUserTypeToEnum(staff.UserType),
		})
	}
	return list, nil
}

func (s *StaffService) GetForEdit(ctx context.Context, staffID string) (models.AdminStaffEditItem, error) {
	decodedID := s.encoder.Decode(staffID)
	if decodedID == encode.INVALID {
		return models.AdminStaffEditItem{}, errs.ErrDecode
	}

	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedID)
	if err != nil {
		return models.AdminStaffEditItem{}, err
	}

	return models.AdminStaffEditItem{
		ID:              staffID,
		FullName:        utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Status:          enums.ParseStaffStatusToEnum(staff.Status),
		Position:        staff.Position,
		TimeInSchedule:  staff.TimeInSchedule.String,
		TimeOutSchedule: staff.TimeOutSchedule.String,
		RequireInShop:   staff.RequireInShop,
	}, nil
}

func (s *StaffService) UpdateEmployment(ctx context.Context, params UpdateEmploymentParams) error {
	if !params.Status.IsValid() {
		return errs.ErrInvalidInput
	}

	if params.Position == "" || params.TimeInSchedule == "" || params.TimeOutSchedule == "" {
		return errs.ErrMissingField
	}

	if _, err := time.Parse(constants.TimeLayoutHHMM, params.TimeInSchedule); err != nil {
		return errs.ErrInvalidFormat
	}

	if _, err := time.Parse(constants.TimeLayoutHHMM, params.TimeOutSchedule); err != nil {
		return errs.ErrInvalidFormat
	}

	decodedID := s.encoder.Decode(params.ID)
	if decodedID == encode.INVALID {
		return errs.ErrDecode
	}

	_, err := s.dbRW.GetQueries().UpdateStaffEmployment(ctx, queries.UpdateStaffEmploymentParams{
		Status:          params.Status.String(),
		Position:        params.Position,
		TimeInSchedule:  sql.NullString{Valid: true, String: params.TimeInSchedule},
		TimeOutSchedule: sql.NullString{Valid: true, String: params.TimeOutSchedule},
		RequireInShop:   params.RequireInShop,
		ID:              decodedID,
	})
	return err
}

func (s *StaffService) GetAllForAdmin(ctx context.Context, search string) ([]queries.GetAllStaffsForAdminRow, error) {
	return s.dbRO.GetQueries().GetAllStaffsForAdmin(ctx, search)
}

func (s *StaffService) GetCurrentStaff(ctx context.Context, staffID string) (queries.GetStaffByIDRow, error) {
	decodedID := s.encoder.Decode(staffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedID)
	return staff, err
}

func (s *StaffService) BuildProfile(staff queries.GetStaffByIDRow) models.AdminStaffProfile {
	return models.AdminStaffProfile{
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

func (s *StaffService) GetAttendanceByDate(ctx context.Context, staffID int64, date string) (queries.GetStaffAttendanceByDateRow, error) {
	return s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: date,
	})
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

func (s *StaffService) ID() string {
	return "Staff"
}

func (s *StaffService) Log() {
	logs.Log().Info("[StaffService] Loaded")
}

var _ IService = (*StaffService)(nil)
