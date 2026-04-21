package services

import "cchoice/internal/enums"

type ProductInventory struct {
	ID        string
	ProductID string
	Stocks    int64
	StocksIn  enums.StocksIn
	CreatedAt string
	UpdatedAt string
}
