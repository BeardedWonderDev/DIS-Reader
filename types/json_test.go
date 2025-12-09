package types

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestSendJSONResponseWritesStatusBodyAndHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	body := map[string]string{"ok": "true"}
	SendJSONResponse(rr, body)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected json content type, got %s", ct)
	}
	if !strings.Contains(rr.Body.String(), `"ok":"true"`) {
		t.Fatalf("expected body to contain encoded json, got %s", rr.Body.String())
	}
}

func TestSendErrorResponseUsesAPIStatusAndBody(t *testing.T) {
	rr := httptest.NewRecorder()
	err := NewForbiddenError("denied")
	SendErrorResponse(rr, err)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"code":"forbidden"`) {
		t.Fatalf("expected json body with code, got %s", rr.Body.String())
	}
}

type demoPayload struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"min=1"`
}

func TestParseJSONBodyValidatesAndParses(t *testing.T) {
	logger := slog.Default()
	body := bytes.NewBufferString(`{"name":"alice","age":5}`)
	var payload demoPayload
	if err := ParseJSONBody(logger, body, &payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if payload.Name != "alice" || payload.Age != 5 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestParseJSONBodyRejectsInvalidJSON(t *testing.T) {
	logger := slog.Default()
	body := bytes.NewBufferString(`{"name":123}`)
	var payload demoPayload
	if err := ParseJSONBody(logger, body, &payload); err == nil {
		t.Fatalf("expected error for type mismatch")
	}
}

func TestPrimitiveTypeToDisplayName(t *testing.T) {
	cases := map[reflect.Type]string{
		reflect.TypeOf(true):         "true or false",
		reflect.TypeOf(""):           "a string",
		reflect.TypeOf(int32(0)):     "a number",
		reflect.TypeOf(uint64(0)):    "a number",
		reflect.TypeOf(float64(0.1)): "a decimal",
		reflect.TypeOf(int8(0)):      "a number",
		reflect.TypeOf(uint8(0)):     "a number",
		reflect.TypeOf(uintptr(0)):   "a number",
		reflect.TypeOf(int16(0)):     "a number",
		reflect.TypeOf(int64(0)):     "a number",
		reflect.TypeOf(uint16(0)):    "a number",
		reflect.TypeOf(uint32(0)):    "a number",
		reflect.TypeOf(complex64(0)): "type complex64",
	}
	for typ, want := range cases {
		if got := primitiveTypeToDisplayName(typ); got != want {
			t.Fatalf("type %v: expected %q, got %q", typ, want, got)
		}
	}
}

func TestParseJSONBytesTypeError(t *testing.T) {
	var dest struct {
		Count int `json:"count"`
	}
	body := []byte(`{"count":"not-a-number"}`)
	err := ParseJSONBytes(body, &dest)
	if err == nil {
		t.Fatalf("expected error for type mismatch")
	}
	if _, ok := err.(*InvalidParameterError); !ok {
		t.Fatalf("expected InvalidParameterError, got %T", err)
	}
}
