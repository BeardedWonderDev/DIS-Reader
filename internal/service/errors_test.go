package service

import (
	"net/http"
	"strings"
	"testing"
)

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		tag      string
		code     string
		status   int
		contains string
	}{
		{"internal", NewInternalError("boom"), "InternalError", ErrorInternalError, http.StatusInternalServerError, "boom"},
		{"invalidRequest", NewInvalidRequestError("bad"), "InvalidRequestError", ErrorInvalidRequest, http.StatusBadRequest, "bad"},
		{"invalidParam", NewInvalidParameterError("field", "wrong"), "InvalidParameterError", ErrorInvalidParameter, http.StatusBadRequest, "field"},
		{"missingParam", NewMissingRequiredParameterError("id"), "MissingRequiredParameterError", ErrorMissingRequiredParameter, http.StatusBadRequest, "id"},
		{"notFound", NewRecordNotFoundError("unit", "u1"), "RecordNotFoundError", ErrorNotFound, http.StatusNotFound, "unit"},
		{"duplicate", NewDuplicateRecordError("unit", "u1", "reason"), "DuplicateRecordError", ErrorDuplicateRecord, http.StatusBadRequest, "reason"},
		{"tokenExpired", NewTokenExpiredError(), "TokenExpiredError", ErrorTokenExpired, http.StatusUnauthorized, "expired"},
		{"tooMany", NewTooManyRequestsError(), "TooManyRequestsError", ErrorTooManyRequests, http.StatusTooManyRequests, "Too many"},
		{"unauthorized", NewUnauthorizedError("nope"), "UnauthorizedError", ErrorUnauthorized, http.StatusUnauthorized, "nope"},
		{"unknownOrigin", NewUnknownOriginError("evil.com"), "UnknownOriginError", ErrorUnknownOrigin, http.StatusForbidden, "evil.com"},
		{"forbidden", NewForbiddenError("no access"), "ForbiddenError", ErrorForbidden, http.StatusForbidden, "no access"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ge, ok := tc.err.(Error)
			if !ok {
				t.Fatalf("expected Error interface")
			}
			if ge.GetTag() != tc.tag {
				t.Fatalf("tag: got %s want %s", ge.GetTag(), tc.tag)
			}
			if ge.GetStatus() != tc.status {
				t.Fatalf("status: got %d want %d", ge.GetStatus(), tc.status)
			}
			if tc.contains != "" && !strings.Contains(tc.err.Error(), tc.contains) {
				t.Fatalf("message %q does not contain %q", tc.err.Error(), tc.contains)
			}
			if g, ok := tc.err.(*genericError); ok && g.Code != tc.code {
				t.Fatalf("code: got %s want %s", g.Code, tc.code)
			}
		})
	}
}
