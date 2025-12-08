package parts

import (
	"testing"
)

func TestToPartInventorySpecTrimsAndMapsFields(t *testing.T) {
	p := &PartInventory{
		Division:    " 01 ",
		VendorCode:  " V1 ",
		PartNumber:  " P123 ",
		Description: " Desc ",
		BinLocation: " BIN ",
		Remaining: map[string]interface{}{
			"PIMC01": "5",
			"PIMC02": float64(3),
			"PIMY1C": "7",
			"PIMY1P": "9",
		},
	}

	spec := p.ToPartInventorySpec()
	if spec.Division != "01" || spec.PartNumber != "P123" || spec.Description != "Desc" {
		t.Fatalf("expected trimmed strings in spec, got %+v", spec)
	}
	if len(spec.MonthlyUsage) != 108 || spec.MonthlyUsage[0] != 5 || spec.MonthlyUsage[1] != 3 {
		t.Fatalf("expected monthly usage to include remaining values")
	}
	if len(spec.YearUsageQty) != 9 || spec.YearUsageQty[1] != 7 {
		t.Fatalf("expected yearly qty mapped from remaining")
	}
	if len(spec.YearUsageValue) != 9 || spec.YearUsageValue[1] != 9 {
		t.Fatalf("expected yearly value mapped from remaining")
	}
}

func TestToIntHelpersHandleMixedTypes(t *testing.T) {
	if got := toInt32(" 12 "); got != 12 {
		t.Fatalf("expected parse string to int32, got %d", got)
	}
	if got := toInt32("bad"); got != 0 {
		t.Fatalf("expected invalid string to yield 0, got %d", got)
	}
	if got := toInt64(float64(15)); got != 15 {
		t.Fatalf("expected float64 to int64 conversion")
	}
	if got := toInt64(nil); got != 0 {
		t.Fatalf("expected nil to return 0")
	}
}
