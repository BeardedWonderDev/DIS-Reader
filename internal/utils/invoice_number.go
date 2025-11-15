package utils

import "strings"

// SanitizeInvoiceNumber trims whitespace and removes trailing artifacts consistently.
func SanitizeInvoiceNumber(val string) string {
	trimmed := strings.TrimSpace(val)
	return strings.TrimSuffix(trimmed, " #*")
}
