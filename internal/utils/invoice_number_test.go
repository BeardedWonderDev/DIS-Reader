package utils

import "testing"

func TestSanitizeInvoiceNumber(t *testing.T) {
	got := SanitizeInvoiceNumber(" 12345 #* ")
	if got != "12345" {
		t.Fatalf("expected trimmed invoice number, got %q", got)
	}

	if SanitizeInvoiceNumber("ABC123") != "ABC123" {
		t.Fatalf("expected unchanged when no suffix")
	}
}
