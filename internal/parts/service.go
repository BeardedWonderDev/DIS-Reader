package parts

import (
	"github.com/BeardedWonderDev/DIS-Reader/internal/utils"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Example function for parsing a Part from a map[string]string (such as CSV row)
func NewPartFromRow(row map[string]string) (*types.Part, error) {
	dateAdded, _ := utils.ParseDate(row["PIMDAD"])
	dateLastChg, _ := utils.ParseDate(row["PIMDLC"])
	dateLastSale, _ := utils.ParseDate(row["PIMDLS"])
	lastPurchase, _ := utils.ParseDate(row["PIMPUP"])
	lastRcvDate, _ := utils.ParseDate(row["PIMLRD"])

	return &types.Part{
		Division:     row["PIMDIV"],
		Vendor:       row["PIMVEN"],
		PartNumber:   row["PIMPRT"],
		Description:  row["PIMDES"],
		Bin:          row["PIMBIN"],
		Class:        row["PIMCLS"],
		Active:       row["PIMACT"],
		Taxable:      row["PIMTAX"],
		DateAdded:    dateAdded,
		DateLastChg:  dateLastChg,
		DateLastSale: dateLastSale,
		Weight:       utils.ParseFloat(row["PIMWHT"]),
		PkgQty:       utils.ParseInt(row["PIMPKG"]),
		QtyRequired:  utils.ParseInt(row["PIMJOB"]),
		OnHand:       utils.ParseInt(row["PIMONH"]),
		ReservedRO:   utils.ParseInt(row["PIMRSA"]),
		ReservedCT:   utils.ParseInt(row["PIMRWO"]),
		Available:    utils.ParseInt(row["PIMAVL"]),
		LastPurchase: lastPurchase,
		AvgCost:      utils.ParseFloat(row["PIMAVG"]),
		DealerList:   utils.ParseFloat(row["PIMCBS"]),
		DlrNetPrice:  utils.ParseFloat(row["PIMCVA"]),
		IntPrice:     utils.ParseFloat(row["PIMIVA"]),
		SuggList:     utils.ParseFloat(row["PIMSVA"]),
		ListPrice:    utils.ParseFloat(row["PIMLVA"]),
		DeliveryCode: row["PIMDLV"],
		OrderCode:    row["PIMORC"],
		MinQty:       utils.ParseInt(row["PIMMIN"]),
		MaxQty:       utils.ParseInt(row["PIMMAX"]),
		OrderQty:     utils.ParseInt(row["PIMODQ"]),
		LastRcvDate:  lastRcvDate,
		LastRcvQty:   utils.ParseInt(row["PIMLRQ"]),
		DealerOrder:  row["PIMPSO"],
		APRCode:      row["PIMAPR"],
		APRSource:    row["PIMAPC"],
		Bulk:         utils.ParseInt(row["PIM9BL"]),
		Comments:     row["PIMCMT"],
	}, nil
}
