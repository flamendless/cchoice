package services

import (
	"context"
	"database/sql"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type StaffService struct {
	encoder encode.IEncode
	dbRO    database.Service
	dbRW    database.Service
}

type UpdateProfileParams struct {
	ID         string
	FirstName  string
	MiddleName string
	LastName   string
	MobileNo   string
	Birthdate  string
	DateHired  string
}

func NewStaffService(
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
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
