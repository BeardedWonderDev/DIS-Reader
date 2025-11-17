package parts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// PartInventory mirrors the AS/400 parts master row and exposes helpers
// to translate it into exported specs.
type PartInventory struct {
	Division            string                 `mapstructure:"PIMDIV"`
	VendorCode          string                 `mapstructure:"PIMVEN"`
	PartNumber          string                 `mapstructure:"PIMPRT"`
	Description         string                 `mapstructure:"PIMDES"`
	BinLocation         string                 `mapstructure:"PIMBIN"`
	QuickCode           string                 `mapstructure:"PIMQKC"`
	ItemClass           string                 `mapstructure:"PIMCLS"`
	ActiveFlag          string                 `mapstructure:"PIMACT"`
	TaxCode             string                 `mapstructure:"PIMTAX"`
	DateAddedYYYYMM     int32                  `mapstructure:"PIMDAD"`
	DateLastCountYYYYMM int32                  `mapstructure:"PIMDLC"`
	DateLastSaleYYYYMM  int32                  `mapstructure:"PIMDLS"`
	UnitWeight          float64                `mapstructure:"PIMWHT"`
	BackorderFlag       string                 `mapstructure:"PIMQBK"`
	PackQty             int32                  `mapstructure:"PIMPKG"`
	OrderMultiplier     int32                  `mapstructure:"PIMMUL"`
	JobCode             int32                  `mapstructure:"PIMJOB"`
	OnHandQty           int32                  `mapstructure:"PIMONH"`
	ReservedSA          int32                  `mapstructure:"PIMRSA"`
	ReservedWO          int32                  `mapstructure:"PIMRWO"`
	AvailableQty        int32                  `mapstructure:"PIMAVL"`
	ReturnableFlag      string                 `mapstructure:"PIMRET"`
	DateLastPriceUpdate int32                  `mapstructure:"PIMPUP"`
	AvgCost             float64                `mapstructure:"PIMAVG"`
	Price1Basis         string                 `mapstructure:"PIMCBS"`
	Price1Value         float64                `mapstructure:"PIMCVA"`
	Price2Basis         string                 `mapstructure:"PIMIBS"`
	Price2Value         float64                `mapstructure:"PIMIVA"`
	Price3Basis         string                 `mapstructure:"PIMSBS"`
	Price3Value         float64                `mapstructure:"PIMSVA"`
	Price4Basis         string                 `mapstructure:"PIMLBS"`
	Price4Value         float64                `mapstructure:"PIMLVA"`
	DeliverCode         string                 `mapstructure:"PIMDLV"`
	OrderCode           string                 `mapstructure:"PIMORC"`
	MinQty              int32                  `mapstructure:"PIMMIN"`
	MaxQty              int32                  `mapstructure:"PIMMAX"`
	OrderQty            int32                  `mapstructure:"PIMODQ"`
	LastReceiptDate     int32                  `mapstructure:"PIMLRD"`
	LastReceiptQty      int32                  `mapstructure:"PIMLRQ"`
	StockFlag           string                 `mapstructure:"PIMSTK"`
	OpenOnSales         int32                  `mapstructure:"PIMOOS"`
	OpenOnPO            int32                  `mapstructure:"PIMOOP"`
	OpenOnRO            int32                  `mapstructure:"PIMOOR"`
	OpenOnReq           int32                  `mapstructure:"PIMOOA"`
	Peak12MoUsage       int32                  `mapstructure:"PIMHYM"`
	Comment             string                 `mapstructure:"PIMCMT"`
	CycleCountCode      string                 `mapstructure:"PIMCCD"`
	CorePartNumber      string                 `mapstructure:"PIMCOR"`
	CoreQty             int32                  `mapstructure:"PIMCQT"`
	CorePrice           float64                `mapstructure:"PIMCPR"`
	Sup1Vendor          string                 `mapstructure:"PIMSSV"`
	Sup1Part            string                 `mapstructure:"PIMSS#"`
	Sup1Price           float64                `mapstructure:"PIMSSP"`
	Sup2Vendor          string                 `mapstructure:"PIMS2V"`
	Sup2Part            string                 `mapstructure:"PIMS2#"`
	Sup2Price           float64                `mapstructure:"PIMS2P"`
	Sup3Vendor          string                 `mapstructure:"PIMS3V"`
	Sup3Part            string                 `mapstructure:"PIMS3#"`
	Sup3Price           float64                `mapstructure:"PIMS3P"`
	RebuildIndicator    string                 `mapstructure:"PIMRBI"`
	RebuildVendor       string                 `mapstructure:"PIMRBV"`
	RebuildPart         string                 `mapstructure:"PIMRB#"`
	RebuildQtyFactor    int32                  `mapstructure:"PIMRBQ"`
	PriceScheduleCode   string                 `mapstructure:"PIMPSO"`
	AutoPricingRule     string                 `mapstructure:"PIMAPR"`
	AutoPricingClass    string                 `mapstructure:"PIMAPC"`
	AutoPricingSource   string                 `mapstructure:"PIMAPS"`
	AutoPricingQtyBrk   int32                  `mapstructure:"PIMAPQ"`
	AltAvailQty         float64                `mapstructure:"PIMAVB"`
	AltOnHandQty        float64                `mapstructure:"PIMONB"`
	AltUomUnits         int32                  `mapstructure:"PIM9UN"`
	AltUomCode          string                 `mapstructure:"PIM9UD"`
	AltUomMultiple      int32                  `mapstructure:"PIM9UM"`
	AltUomBackorderFlg  string                 `mapstructure:"PIM9BL"`
	AltUomFreightFlg    string                 `mapstructure:"PIM9FR"`
	Message1            int32                  `mapstructure:"PIMSG1"`
	Message2            int32                  `mapstructure:"PIMSG2"`
	Message3            int32                  `mapstructure:"PIMSG3"`
	Remaining           map[string]interface{} `mapstructure:",remain"`
}

// ToPartInventorySpec converts the repository model into an exported spec.
func (p *PartInventory) ToPartInventorySpec() *types.PartInventorySpec {
	return &types.PartInventorySpec{
		Division:            strings.TrimSpace(p.Division),
		VendorCode:          strings.TrimSpace(p.VendorCode),
		PartNumber:          strings.TrimSpace(p.PartNumber),
		Description:         strings.TrimSpace(p.Description),
		BinLocation:         strings.TrimSpace(p.BinLocation),
		QuickCode:           strings.TrimSpace(p.QuickCode),
		ItemClass:           strings.TrimSpace(p.ItemClass),
		ActiveFlag:          strings.TrimSpace(p.ActiveFlag),
		TaxCode:             strings.TrimSpace(p.TaxCode),
		DateAddedYYYYMM:     p.DateAddedYYYYMM,
		DateLastCountYYYYMM: p.DateLastCountYYYYMM,
		DateLastSaleYYYYMM:  p.DateLastSaleYYYYMM,
		UnitWeight:          p.UnitWeight,
		BackorderFlag:       strings.TrimSpace(p.BackorderFlag),
		PackQty:             p.PackQty,
		OrderMultiplier:     p.OrderMultiplier,
		JobCode:             p.JobCode,
		OnHandQty:           p.OnHandQty,
		ReservedSA:          p.ReservedSA,
		ReservedWO:          p.ReservedWO,
		AvailableQty:        p.AvailableQty,
		ReturnableFlag:      strings.TrimSpace(p.ReturnableFlag),
		DateLastPriceUpdate: p.DateLastPriceUpdate,
		AvgCost:             p.AvgCost,
		Price1Basis:         strings.TrimSpace(p.Price1Basis),
		Price1Value:         p.Price1Value,
		Price2Basis:         strings.TrimSpace(p.Price2Basis),
		Price2Value:         p.Price2Value,
		Price3Basis:         strings.TrimSpace(p.Price3Basis),
		Price3Value:         p.Price3Value,
		Price4Basis:         strings.TrimSpace(p.Price4Basis),
		Price4Value:         p.Price4Value,
		DeliverCode:         strings.TrimSpace(p.DeliverCode),
		OrderCode:           strings.TrimSpace(p.OrderCode),
		MinQty:              p.MinQty,
		MaxQty:              p.MaxQty,
		OrderQty:            p.OrderQty,
		LastReceiptDate:     p.LastReceiptDate,
		LastReceiptQty:      p.LastReceiptQty,
		StockFlag:           strings.TrimSpace(p.StockFlag),
		OpenOnSales:         p.OpenOnSales,
		OpenOnPO:            p.OpenOnPO,
		OpenOnRO:            p.OpenOnRO,
		OpenOnReq:           p.OpenOnReq,
		Peak12MoUsage:       p.Peak12MoUsage,
		MonthlyUsage:        p.extractSequentialInt32("PIMC", 108),
		MonthlyPurchases:    p.extractSequentialInt32("PIMP", 108),
		YearUsageQty:        p.extractYearlyInt32("C"),
		YearUsageValue:      p.extractYearlyInt64("P"),
		Comment:             strings.TrimSpace(p.Comment),
		CycleCountCode:      strings.TrimSpace(p.CycleCountCode),
		CorePartNumber:      strings.TrimSpace(p.CorePartNumber),
		CoreQty:             p.CoreQty,
		CorePrice:           p.CorePrice,
		Sup1Vendor:          strings.TrimSpace(p.Sup1Vendor),
		Sup1Part:            strings.TrimSpace(p.Sup1Part),
		Sup1Price:           p.Sup1Price,
		Sup2Vendor:          strings.TrimSpace(p.Sup2Vendor),
		Sup2Part:            strings.TrimSpace(p.Sup2Part),
		Sup2Price:           p.Sup2Price,
		Sup3Vendor:          strings.TrimSpace(p.Sup3Vendor),
		Sup3Part:            strings.TrimSpace(p.Sup3Part),
		Sup3Price:           p.Sup3Price,
		RebuildIndicator:    strings.TrimSpace(p.RebuildIndicator),
		RebuildVendor:       strings.TrimSpace(p.RebuildVendor),
		RebuildPart:         strings.TrimSpace(p.RebuildPart),
		RebuildQtyFactor:    p.RebuildQtyFactor,
		PriceScheduleCode:   strings.TrimSpace(p.PriceScheduleCode),
		AutoPricingRule:     strings.TrimSpace(p.AutoPricingRule),
		AutoPricingClass:    strings.TrimSpace(p.AutoPricingClass),
		AutoPricingSource:   strings.TrimSpace(p.AutoPricingSource),
		AutoPricingQtyBrk:   p.AutoPricingQtyBrk,
		AltAvailQty:         p.AltAvailQty,
		AltOnHandQty:        p.AltOnHandQty,
		AltUomUnits:         p.AltUomUnits,
		AltUomCode:          strings.TrimSpace(p.AltUomCode),
		AltUomMultiple:      p.AltUomMultiple,
		AltUomBackorderFlag: strings.TrimSpace(p.AltUomBackorderFlg),
		AltUomFreightFlag:   strings.TrimSpace(p.AltUomFreightFlg),
		Message1:            p.Message1,
		Message2:            p.Message2,
		Message3:            p.Message3,
	}
}

func (p *PartInventory) extractSequentialInt32(prefix string, count int) []int32 {
	seq := make([]int32, count)
	for i := 1; i <= count; i++ {
		key := fmt.Sprintf("%s%02d", prefix, i)
		if val, ok := p.Remaining[key]; ok {
			seq[i-1] = toInt32(val)
		}
	}
	return seq
}

func (p *PartInventory) extractYearlyInt32(suffix string) []int32 {
	seq := make([]int32, 9)
	for i := range seq {
		key := fmt.Sprintf("PIMY%d%s", i, suffix)
		if val, ok := p.Remaining[key]; ok {
			seq[i] = toInt32(val)
		}
	}
	return seq
}

func (p *PartInventory) extractYearlyInt64(suffix string) []int64 {
	seq := make([]int64, 9)
	for i := range seq {
		key := fmt.Sprintf("PIMY%d%s", i, suffix)
		if val, ok := p.Remaining[key]; ok {
			seq[i] = toInt64(val)
		}
	}
	return seq
}

func toInt32(val interface{}) int32 {
	switch v := val.(type) {
	case nil:
		return 0
	case int:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	case float32:
		return int32(v)
	case float64:
		return int32(v)
	case string:
		if v == "" {
			return 0
		}
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return int32(n)
	default:
		return 0
	}
}

func toInt64(val interface{}) int64 {
	switch v := val.(type) {
	case nil:
		return 0
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		if v == "" {
			return 0
		}
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return int64(n)
	default:
		return 0
	}
}
