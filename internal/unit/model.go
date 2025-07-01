package unit

import (
	"strings"
	"time"
)

type Model interface {
	GetUnitID() string
	GetYear() int
	GetMake() string
	GetModel() string
	GetDescription() string
	GetSerial() *string
	GetEngine() *string
	GetStatus() string
	GetProductCode() *string
	GetIsNew() bool
	GetColor() *string
	GetHorsePower() *int
	GetEngineHours() *float64
	GetLocation() string
	GetAccount() string
	GetCreatedAt() time.Time
	GetSoldAt() time.Time
	GetInvoiceNumber() *string
	GetSoldAccount() *string
	GetRevenueAmount() *float64
	GetCost() float64
	GetSoldTo() *string
	GetSoldBy() *string
	GetSoldByName() *string
	GetWarrantyCode() *string
	GetTradedOnUnit() *string
	GetGLAccount() *string
	GetSuggestedList() *float64
	GetDealerList() *float64
	GetFlooringDueDate() *time.Time
	GetFlooringAmount() *float64
	GetRentalStartDate() *time.Time
	GetRentalRevenue() *float64
	GetRentalCost() *float64
	GetSoldToName() *string
	GetSoldToAddress() *string
	GetSoldToCity() *string
	GetSoldToZip() *string
	GetPhoneNumber() *string
	ToUnitSpec() *UnitSpec
}

type Unit struct {
	UnitID          string     `mapstruct:"UNIT"`
	Year            int        `mapstruct:"YEAR"`
	Make            string     `mapstruct:"MAKE"`
	Model           string     `mapstruct:"MODEL"`
	Description     string     `mapstruct:"DESCR"`
	Serial          *string    `mapstruct:"SERIAL"`
	Engine          *string    `mapstruct:"ENGINE"`
	Status          string     `mapstruct:"STATUS"`
	ProductCode     *string    `mapstruct:"PRODCT"`
	IsNew           string     `mapstruct:"NEWUSE"`
	Color           *string    `mapstruct:"COLOR"`
	HorsePower      *int       `mapstruct:"HRSPWR"`
	EngineHours     *float64   `mapstruct:"HOUR"`
	Location        string     `mapstruct:"LOCATE"`
	Account         string     `mapstruct:"ACCT#"`
	CreatedAt       time.Time  `mapstruct:"INDATE"`
	SoldAt          time.Time  `mapstruct:"SALEDT"`
	InvoiceNumber   *string    `mapstruct:"INV#"`
	SoldAccount     *string    `mapstruct:"SOLDTO"`
	RevenueAmount   *float64   `mapstruct:"AMOUNT"`
	Cost            float64    `mapstruct:"COST"`
	SoldBy          *string    `mapstruct:"SOLDBY"`
	SoldByName      *string    `mapstruct:"BYNAME"`
	WarrantyCode    *string    `mapstruct:"WARRCD"`
	TradedOnUnit    *string    `mapstruct:"TRADE"`
	GLAccount       *string    `mapstruct:"INACCT"`
	SuggestedList   *float64   `mapstruct:"SLIST"`
	DealerList      *float64   `mapstruct:"DLIST"`
	FlooringDueDate *time.Time `mapstruct:"FLRDUE"`
	FlooringAmount  *float64   `mapstruct:"FLRAMT"`
	RentalStartDate *time.Time `mapstruct:"RENTDT"`
	RentalRevenue   *float64   `mapstruct:"RENTRV"`
	RentalCost      *float64   `mapstruct:"RENTCS"`
	SoldToName      *string    `mapstruct:"TONAME"`
	SoldToAddress   *string    `mapstruct:"TOADRS"`
	SoldToCity      *string    `mapstruct:"TOCITY"`
	SoldToZip       *string    `mapstruct:"TOZIP"`
	PhoneNumber     *string    `mapstruct:"PHONE"`
}

func (u Unit) GetUnitID() string              { return u.UnitID }
func (u Unit) GetYear() int                   { return u.Year }
func (u Unit) GetMake() string                { return u.Make }
func (u Unit) GetModel() string               { return u.Model }
func (u Unit) GetDescription() string         { return u.Description }
func (u Unit) GetSerial() *string             { return u.Serial }
func (u Unit) GetEngine() *string             { return u.Engine }
func (u Unit) GetStatus() string              { return u.Status }
func (u Unit) GetProductCode() *string        { return u.ProductCode }
func (u Unit) GetIsNew() bool                 { return strings.ToUpper(u.IsNew) == "N" }
func (u Unit) GetColor() *string              { return u.Color }
func (u Unit) GetHorsePower() *int            { return u.HorsePower }
func (u Unit) GetEngineHours() *float64       { return u.EngineHours }
func (u Unit) GetLocation() string            { return u.Location }
func (u Unit) GetAccount() string             { return u.Account }
func (u Unit) GetCreatedAt() time.Time        { return u.CreatedAt }
func (u Unit) GetSoldAt() time.Time           { return u.SoldAt }
func (u Unit) GetInvoiceNumber() *string      { return u.InvoiceNumber }
func (u Unit) GetSoldAccount() *string        { return u.SoldAccount }
func (u Unit) GetRevenueAmount() *float64     { return u.RevenueAmount }
func (u Unit) GetCost() float64               { return u.Cost }
func (u Unit) GetSoldTo() *string             { return u.SoldAccount }
func (u Unit) GetSoldBy() *string             { return u.SoldBy }
func (u Unit) GetSoldByName() *string         { return u.SoldByName }
func (u Unit) GetWarrantyCode() *string       { return u.WarrantyCode }
func (u Unit) GetTradedOnUnit() *string       { return u.TradedOnUnit }
func (u Unit) GetGLAccount() *string          { return u.GLAccount }
func (u Unit) GetSuggestedList() *float64     { return u.SuggestedList }
func (u Unit) GetDealerList() *float64        { return u.DealerList }
func (u Unit) GetFlooringDueDate() *time.Time { return u.FlooringDueDate }
func (u Unit) GetFlooringAmount() *float64    { return u.FlooringAmount }
func (u Unit) GetRentalStartDate() *time.Time { return u.RentalStartDate }
func (u Unit) GetRentalRevenue() *float64     { return u.RentalRevenue }
func (u Unit) GetRentalCost() *float64        { return u.RentalCost }
func (u Unit) GetSoldToName() *string         { return u.SoldToName }
func (u Unit) GetSoldToAddress() *string      { return u.SoldToAddress }
func (u Unit) GetSoldToCity() *string         { return u.SoldToCity }
func (u Unit) GetSoldToZip() *string          { return u.SoldToZip }
func (u Unit) GetPhoneNumber() *string        { return u.PhoneNumber }
func (u Unit) ToUnitSpec() *UnitSpec {
	return &UnitSpec{
		UnitID:        u.UnitID,
		Year:          u.Year,
		Make:          u.Make,
		Model:         u.Model,
		Description:   u.Description,
		Serial:        u.Serial,
		Engine:        u.Engine,
		Status:        u.Status,
		ProductCode:   u.ProductCode,
		New:           u.GetIsNew(),
		Color:         u.Color,
		HorsePower:    u.HorsePower,
		EngineHours:   u.EngineHours,
		Location:      u.Location,
		Account:       u.Account,
		CreatedAt:     u.CreatedAt,
		SoldAt:        u.SoldAt,
		InvoiceNumber: u.InvoiceNumber,
		SoldTo:        u.SoldAccount,
		RevenueAmount: u.RevenueAmount,
		Cost:          u.Cost,
		SoldBy:        u.SoldBy,
		SoldByName:    u.SoldByName,
	}
}
