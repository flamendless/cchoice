package services

import "time"

type CreateCpointParams struct {
	ExpiresAt   *time.Time
	StaffID     string
	CustomerID  string
	ProductSkus []string
	Value       int64
}

type Cpoint struct {
	GeneratedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   *time.Time
	RedeemedAt  *time.Time
	Code        string
	ProductSkus []string
	ID          int64
	CustomerID  int64
	Value       int64
}

type GetRedeemedCpointsWithTotal struct {
	CPoints []Cpoint
	Total   int64
}
