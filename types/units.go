package types

import (
	"context"
	"time"
)

type UnitService interface {
	GetByUnitNumber(ctx context.Context, unitNum string) (*UnitSpec, error)
	ListUnits(ctx context.Context, lp ListParams) ([]*UnitSpec, error)
}

type UnitSpec struct {
	UnitID        string    `json:"unitId" validate:"required"`
	Year          string    `json:"year"`
	Make          string    `json:"make"`
	Model         string    `json:"model"`
	Description   string    `json:"description"`
	Serial        *string   `json:"serial,omitempty"`
	Engine        *string   `json:"engine,omitempty"`
	Status        string    `json:"status"`
	ProductCode   *string   `json:"productCode,omitempty"`
	New           bool      `json:"new"`
	Color         *string   `json:"color,omitempty"`
	HorsePower    *int      `json:"horsePower,omitempty"`
	EngineHours   *float64  `json:"engineHours,omitempty"`
	Location      string    `json:"location"`
	Account       string    `json:"account"`
	CreatedAt     time.Time `json:"createdAt"`
	SoldAt        time.Time `json:"soldAt,omitempty"`
	InvoiceNumber *string   `json:"invoiceNumber,omitempty"`
	SoldTo        *string   `json:"soldTo,omitempty"`
	RevenueAmount *float64  `json:"revenueAmount,omitempty"`
	Cost          float64   `json:"cost,omitempty"`
	SoldBy        *string   `json:"soldBy,omitempty"`
	SoldByName    *string   `json:"soldByName,omitempty"`
}
