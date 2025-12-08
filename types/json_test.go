package types

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"testing"
)

func TestIsArray(t *testing.T) {
	if !IsArray([]byte("[1,2,3]")) {
		t.Fatalf("expected IsArray to detect array")
	}
	if IsArray([]byte("{\"a\":1}")) {
		t.Fatalf("expected IsArray to be false for object")
	}
}

func TestParseJSONBytesTypeError(t *testing.T) {
	var payload struct {
		Count int `json:"count"`
	}
	err := ParseJSONBytes([]byte(`{"count":"oops"}`), &payload)
	var ip *InvalidParameterError
	if !errors.As(err, &ip) {
		t.Fatalf("expected InvalidParameterError, got %v", err)
	}
	if ip.Parameter != "count" {
		t.Fatalf("expected parameter 'count', got %s", ip.Parameter)
	}
}

func TestValidateStructRequiredIfOneOf(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	type req struct {
		Mode  string   `validate:"oneof=a b" json:"mode"`
		Items []string `validate:"required_if_oneof=Mode a" json:"items"`
	}
	r := req{Mode: "a"}
	err := ValidateStruct(logger, &r)
	var missing *MissingRequiredParameterError
	if !errors.As(err, &missing) || missing.Parameter != "items" {
		t.Fatalf("expected MissingRequiredParameterError for items, got %v", err)
	}

	r.Items = []string{"ok"}
	if err := ValidateStruct(logger, &r); err != nil {
		t.Fatalf("expected validation to pass, got %v", err)
	}
}

func TestParseJSONBodyPointerGuard(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	// Non-pointer should return InternalError
	err := ParseJSONBody(logger, bytes.NewBufferString(`{"a":1}`), struct{}{})
	var internal *InternalError
	if !errors.As(err, &internal) {
		t.Fatalf("expected InternalError for non-pointer dest, got %v", err)
	}
}
