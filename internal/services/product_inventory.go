package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

type ProductInventoryService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewProductInventoryService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *ProductInventoryService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &ProductInventoryService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

func (s *ProductInventoryService) GetByProductID(ctx context.Context, productID string) (*ProductInventory, error) {
	decoded := s.encoder.Decode(productID)
	if decoded == encode.INVALID {
		return nil, errs.ErrDecode
	}

	inv, err := s.dbRO.GetQueries().GetProductInventoryByProductID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Join(errs.ErrProductInventory, err)
	}

	return s.mapRowToProductInventory(inv), nil
}

func (s *ProductInventoryService) GetListingForAdmin(ctx context.Context) ([]models.AdminProductInventoryListItem, error) {
	inventories, err := s.dbRO.GetQueries().AdminGetProductInventoriesListing(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrProductInventory, err)
	}

	items := make([]models.AdminProductInventoryListItem, 0, len(inventories))
	for _, inv := range inventories {
		items = append(items, models.AdminProductInventoryListItem{
			ID:            s.encoder.Encode(inv.ID),
			ProductSerial: inv.ProductSerial,
			StocksIn:      enums.ParseStocksInToEnum(inv.StocksIn),
			Stocks:        inv.Stocks,
			UpdatedAt:     inv.UpdatedAt,
		})
	}

	return items, nil
}

func (s *ProductInventoryService) Create(
	ctx context.Context,
	staffID string,
	productID string,
	stocks int64,
	stocksIn enums.StocksIn,
) (string, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionCreate,
			constants.ModuleProductInventories,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("[ProductInventoryService] create log", zap.Error(err))
		}
	}()

	productDBID := s.encoder.Decode(productID)
	if productDBID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return "", errs.ErrDecode
	}

	_, err := s.dbRO.GetQueries().GetProductInventoryByProductID(ctx, productDBID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		result = err.Error()
		return "", errors.Join(errs.ErrProductInventory, err)
	}
	if err == nil {
		result = "inventory already exists"
		return "", errs.ErrProductInventory
	}

	id, err := s.dbRW.GetQueries().CreateProductInventory(ctx, queries.CreateProductInventoryParams{
		ProductID: productDBID,
		Stocks:    stocks,
		StocksIn:  stocksIn.String(),
	})
	if err != nil {
		result = err.Error()
		return "", errors.Join(errs.ErrProductInventory, err)
	}

	inventoryID := s.encoder.Encode(id)
	result = fmt.Sprintf("success. ID '%s'", inventoryID)
	return inventoryID, nil
}

func (s *ProductInventoryService) SetQty(
	ctx context.Context,
	staffID string,
	productID string,
	qty int64,
	stocksIn enums.StocksIn,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionUpdate,
			constants.ModuleProductInventories,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("[ProductInventoryService] set qty log", zap.Error(err))
		}
	}()

	productDBID := s.encoder.Decode(productID)
	if productDBID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	if err := s.dbRW.GetQueries().UpdateProductInventory(ctx, queries.UpdateProductInventoryParams{
		ProductID: productDBID,
		StocksIn:  stocksIn.String(),
		Stocks:    qty,
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrProductInventory, err)
	}

	result = fmt.Sprintf("success. product ID '%s'", productID)
	return nil
}

func (s *ProductInventoryService) mapRowToProductInventory(p queries.TblProductInventory) *ProductInventory {
	return &ProductInventory{
		ID:        s.encoder.Encode(p.ID),
		ProductID: s.encoder.Encode(p.ProductID),
		Stocks:    p.Stocks,
		StocksIn:  enums.ParseStocksInToEnum(p.StocksIn),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func (s *ProductInventoryService) ID() string {
	return "ProductInventory"
}

func (s *ProductInventoryService) Log() {
	logs.Log().Info("[ProductInventoryService] Loaded")
}

var _ IService = (*ProductInventoryService)(nil)
