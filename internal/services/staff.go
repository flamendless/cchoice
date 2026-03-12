package services

import (
	"context"

	"cchoice/cmd/web/models"
	"cchoice/internal/database"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
)

type StaffService struct {
	dbRO database.Service
}

func NewStaffService(dbRO database.Service) *StaffService {
	return &StaffService{dbRO: dbRO}
}

func (s *StaffService) GetByID(ctx context.Context, staffID int64) (models.AdminStaffProfile, error) {
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
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
