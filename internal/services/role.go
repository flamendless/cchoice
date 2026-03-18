package services

import (
	"context"

	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
)

type RoleService struct {
	encoder encode.IEncode
	dbRO    database.Service
}

func NewRoleService(
	encoder encode.IEncode,
	dbRO database.Service,
) *RoleService {
	return &RoleService{
		encoder: encoder,
		dbRO:    dbRO,
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
