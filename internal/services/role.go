package services

import (
	"context"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
)

type RoleService struct {
	encoder encode.IEncode
	dbRO    database.Service
	dbRW    database.Service
}

func NewRoleService(
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
) *RoleService {
	return &RoleService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
	}
}

func (s *RoleService) GetByStaffID(ctx context.Context, staffID string) ([]enums.StaffRole, error) {
	decodedID := s.encoder.Decode(staffID)
	dbRoles, err := s.dbRO.GetQueries().GetStaffRolesByStaffID(ctx, decodedID)
	if err != nil {
		return nil, err
	}

	roles := make([]enums.StaffRole, 0, len(dbRoles))
	for _, roleStr := range dbRoles {
		role := enums.ParseStaffRoleToEnum(roleStr)
		if role.IsValid() {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (s *RoleService) AddRole(ctx context.Context, staffID string, role enums.StaffRole) error {
	if !role.IsValid() {
		return errs.ErrInvalidParams
	}

	decodedID := s.encoder.Decode(staffID)
	_, err := s.dbRW.GetQueries().CreateStaffRole(ctx, queries.CreateStaffRoleParams{
		StaffID: decodedID,
		Role:    role.String(),
	})
	return err
}

func (s *RoleService) RemoveRole(ctx context.Context, staffID string, role enums.StaffRole) error {
	if !role.IsValid() {
		return errs.ErrInvalidParams
	}

	decodedID := s.encoder.Decode(staffID)
	_, err := s.dbRW.GetQueries().DeleteStaffRole(ctx, queries.DeleteStaffRoleParams{
		StaffID: decodedID,
		Role:    role.String(),
	})
	return err
}
