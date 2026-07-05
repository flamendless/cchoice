package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/orderhistory"

	"go.uber.org/zap"
)

type OrderAdminManageData struct {
	ID             int64
	OrderReference string
	Status         enums.OrderStatus
}

type OrderAdminStatusHistoryEntry struct {
	FromStatus string
	ToStatus   string
	StaffName  string
	Notes      string
	CreatedAt  string
}

type OrderAdminTrackData struct {
	ID             int64
	OrderReference string
	CurrentStatus  enums.OrderStatus
	History        []OrderAdminStatusHistoryEntry
	FlowSteps      []enums.OrderStatus
}

func (s *OrderService) GetManageDataForAdmin(ctx context.Context, orderIDStr string) (*OrderAdminManageData, error) {
	decoded := s.encoder.Decode(orderIDStr)
	if decoded == encode.INVALID {
		return nil, errs.ErrDecode
	}

	order, err := s.dbRO.GetQueries().GetOrderByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &OrderAdminManageData{
		ID:             order.ID,
		OrderReference: order.OrderNumber,
		Status:         enums.ParseOrderStatusToEnum(order.Status),
	}, nil
}

func (s *OrderService) GetStatusHistoryForAdmin(ctx context.Context, orderIDStr string) (*OrderAdminTrackData, error) {
	decoded := s.encoder.Decode(orderIDStr)
	if decoded == encode.INVALID {
		return nil, errs.ErrDecode
	}

	order, err := s.dbRO.GetQueries().GetOrderByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	rows, err := s.dbRO.GetQueries().GetOrderStatusHistoryByOrderID(ctx, decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to get order status history: %w", err)
	}

	history := make([]OrderAdminStatusHistoryEntry, 0, len(rows))
	for _, row := range rows {
		history = append(history, mapOrderStatusHistoryEntry(row))
	}

	currentStatus := enums.ParseOrderStatusToEnum(order.Status)

	return &OrderAdminTrackData{
		ID:             order.ID,
		OrderReference: order.OrderNumber,
		CurrentStatus:  currentStatus,
		History:        history,
		FlowSteps:      buildOrderFlowSteps(history),
	}, nil
}

func (s *OrderService) UpdateOrderForAdmin(
	ctx context.Context,
	staffIDStr string,
	orderIDStr string,
	newStatusStr string,
	notes string,
	canUpdateStatus bool,
) error {
	result := "success"
	action := constants.ActionUpdate
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffIDStr, action, constants.ModuleOrders, result, nil); err != nil {
			logs.Log().Warn("[OrderService] update order log", zap.Error(err))
		}
	}()

	decoded := s.encoder.Decode(orderIDStr)
	if decoded == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	newStatus := enums.ParseOrderStatusToEnum(strings.ToUpper(strings.TrimSpace(newStatusStr)))
	if !newStatus.IsValid() {
		result = errs.ErrEnumInvalid.Error()
		return errs.ErrEnumInvalid
	}

	order, err := s.dbRO.GetQueries().GetOrderByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			result = errs.ErrNotFound.Error()
			return errs.ErrNotFound
		}
		result = err.Error()
		return fmt.Errorf("failed to get order: %w", err)
	}

	currentStatus := enums.ParseOrderStatusToEnum(order.Status)
	statusChanged := newStatus != currentStatus
	notesTrimmed := strings.TrimSpace(notes)

	if statusChanged && !canUpdateStatus {
		result = errs.ErrForbidden.Error()
		return errs.ErrForbidden
	}

	if !statusChanged && notesTrimmed == "" {
		result = errs.ErrMissingField.Error()
		return errs.ErrMissingField
	}

	staffID := s.encoder.Decode(staffIDStr)
	if staffID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	staffIDParam := sql.NullInt64{Int64: staffID, Valid: true}
	notesParam := sql.NullString{String: notesTrimmed, Valid: notesTrimmed != ""}

	if statusChanged {
		action = constants.ActionUpdateStatus
		if _, err := s.dbRW.GetQueries().UpdateOrderStatus(ctx, queries.UpdateOrderStatusParams{
			ID:     decoded,
			Status: newStatus.String(),
		}); err != nil {
			result = err.Error()
			return fmt.Errorf("failed to update order status: %w", err)
		}
	}

	fromStatus := sql.NullString{String: currentStatus.String(), Valid: true}
	toStatus := newStatus.String()
	if !statusChanged {
		toStatus = currentStatus.String()
	}

	if err := orderhistory.Record(ctx, s.dbRW, decoded, staffIDParam, fromStatus, toStatus, notesParam); err != nil {
		result = err.Error()
		return err
	}

	result = fmt.Sprintf("success. order '%s'", orderIDStr)
	return nil
}

func mapOrderStatusHistoryEntry(row queries.GetOrderStatusHistoryByOrderIDRow) OrderAdminStatusHistoryEntry {
	fromStatus := "—"
	if row.FromStatus.Valid && row.FromStatus.String != "" {
		fromStatus = row.FromStatus.String
	}

	staffName := "System"
	if row.StaffID.Valid {
		parts := make([]string, 0, 2)
		if row.StaffFirstName.Valid && row.StaffFirstName.String != "" {
			parts = append(parts, row.StaffFirstName.String)
		}
		if row.StaffLastName.Valid && row.StaffLastName.String != "" {
			parts = append(parts, row.StaffLastName.String)
		}
		if len(parts) > 0 {
			staffName = strings.Join(parts, " ")
		} else {
			staffName = "Staff"
		}
	}

	notes := ""
	if row.Notes.Valid {
		notes = row.Notes.String
	}

	return OrderAdminStatusHistoryEntry{
		FromStatus: fromStatus,
		ToStatus:   row.ToStatus,
		StaffName:  staffName,
		Notes:      notes,
		CreatedAt:  row.CreatedAt.Format(constants.DateTimeLayoutISO),
	}
}

func buildOrderFlowSteps(history []OrderAdminStatusHistoryEntry) []enums.OrderStatus {
	steps := make([]enums.OrderStatus, 0, len(history))
	var prev enums.OrderStatus
	for _, entry := range history {
		status := enums.ParseOrderStatusToEnum(entry.ToStatus)
		if !status.IsValid() {
			continue
		}
		if len(steps) == 0 || status != prev {
			steps = append(steps, status)
			prev = status
		}
	}
	return steps
}
