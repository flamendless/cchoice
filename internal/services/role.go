package services

import (
	"context"
	"fmt"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

type RoleService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewRoleService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *RoleService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &RoleService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
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

func (s *RoleService) AddRole(ctx context.Context, actorStaffID, staffID string, role enums.StaffRole) error {
	return s.mutateRole(ctx, actorStaffID, staffID, role, constants.ActionGrant, func() error {
		decodedID := s.encoder.Decode(staffID)
		_, err := s.dbRW.GetQueries().CreateStaffRole(ctx, queries.CreateStaffRoleParams{
			StaffID: decodedID,
			Role:    role.String(),
		})
		return err
	})
}

func (s *RoleService) RemoveRole(ctx context.Context, actorStaffID, staffID string, role enums.StaffRole) error {
	return s.mutateRole(ctx, actorStaffID, staffID, role, constants.ActionRevoke, func() error {
		decodedID := s.encoder.Decode(staffID)
		_, err := s.dbRW.GetQueries().DeleteStaffRole(ctx, queries.DeleteStaffRoleParams{
			StaffID: decodedID,
			Role:    role.String(),
		})
		return err
	})
}

func (s *RoleService) mutateRole(
	ctx context.Context,
	actorStaffID, staffID string,
	role enums.StaffRole,
	action string,
	mutate func() error,
) error {
	if !role.IsValid() {
		return errs.ErrInvalidParams
	}

	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, actorStaffID, action, constants.ModuleStaff, result, nil); err != nil {
			logs.Log().Warn("[RoleService] staff log", zap.Error(err))
		}
	}()

	if err := mutate(); err != nil {
		result = err.Error()
		return err
	}

	result = fmt.Sprintf("success. staff ID '%s' role '%s'", staffID, role.String())
	return nil
}

func (s *RoleService) ID() string {
	return "Role"
}

func (s *RoleService) Log() {
	logs.Log().Info("[RoleService] Loaded")
}

var _ IService = (*RoleService)(nil)
