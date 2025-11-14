package database

import (
	"fmt"
	"strings"
	"testing"
)

func TestAs400DateExprGuardsInvalidDates(t *testing.T) {
	col := "FILEC.IAH.AHPDTE"
	expr := As400DateExpr(col)
	requiredSubstrings := []string{
		"CASE ",
		"MOD(INTEGER",
		fmt.Sprintf("RIGHT('000000'||TRIM(CHAR(%s)),6)", col),
	}
	for _, substr := range requiredSubstrings {
		if !strings.Contains(expr, substr) {
			t.Fatalf("guarded expression missing %q: %s", substr, expr)
		}
	}
}
