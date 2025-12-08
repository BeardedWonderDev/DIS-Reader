package types

import (
	"net/http"
	"testing"
)

func TestErrorConstructorsExposeTagStatusAndMessage(t *testing.T) {
	cases := []struct {
		name          string
		err           Error
		expectTag     string
		expectStatus  int
		expectMessage string
	}{
		{
			name:          "internal",
			err:           NewInternalError("boom"),
			expectTag:     "InternalError",
			expectStatus:  http.StatusInternalServerError,
			expectMessage: "InternalError: boom",
		},
		{
			name:          "invalid request",
			err:           NewInvalidRequestError("nope"),
			expectTag:     "InvalidRequestError",
			expectStatus:  http.StatusBadRequest,
			expectMessage: "InvalidRequestError: nope",
		},
		{
			name:          "invalid parameter",
			err:           NewInvalidParameterError("field", "bad"),
			expectTag:     "InvalidParameterError",
			expectStatus:  http.StatusBadRequest,
			expectMessage: "InvalidParameterError: Invalid parameter field, bad",
		},
		{
			name:          "missing required parameter",
			err:           NewMissingRequiredParameterError("needme"),
			expectTag:     "MissingRequiredParameterError",
			expectStatus:  http.StatusBadRequest,
			expectMessage: "MissingRequiredParameterError: Missing required parameter needme",
		},
		{
			name:          "record not found",
			err:           NewRecordNotFoundError("Thing", "key-1"),
			expectTag:     "RecordNotFoundError",
			expectStatus:  http.StatusNotFound,
			expectMessage: "RecordNotFoundError: Thing key-1 not found",
		},
		{
			name:          "duplicate record",
			err:           NewDuplicateRecordError("Thing", "key-1", "exists"),
			expectTag:     "DuplicateRecordError",
			expectStatus:  http.StatusBadRequest,
			expectMessage: "DuplicateRecordError: Duplicate Thing key-1, exists",
		},
		{
			name:          "token expired",
			err:           NewTokenExpiredError(),
			expectTag:     "TokenExpiredError",
			expectStatus:  http.StatusUnauthorized,
			expectMessage: "TokenExpiredError: Token is expired.",
		},
		{
			name:          "too many requests",
			err:           NewTooManyRequestsError(),
			expectTag:     "TooManyRequestsError",
			expectStatus:  http.StatusTooManyRequests,
			expectMessage: "TooManyRequestsError: Too many requests.",
		},
		{
			name:          "unauthorized",
			err:           NewUnauthorizedError("no token"),
			expectTag:     "UnauthorizedError",
			expectStatus:  http.StatusUnauthorized,
			expectMessage: "UnauthorizedError: no token",
		},
		{
			name:          "unknown origin",
			err:           NewUnknownOriginError("example.com"),
			expectTag:     "UnknownOriginError",
			expectStatus:  http.StatusForbidden,
			expectMessage: "UnknownOriginError: Request originated from an unknown origin example.com. Configure this origin as an allowed origin from the dashboard to allow requests.",
		},
		{
			name:          "forbidden",
			err:           NewForbiddenError("nope"),
			expectTag:     "ForbiddenError",
			expectStatus:  http.StatusForbidden,
			expectMessage: "ForbiddenError: nope",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.GetTag() != tc.expectTag {
				t.Fatalf("tag mismatch: got %q want %q", tc.err.GetTag(), tc.expectTag)
			}
			if tc.err.GetStatus() != tc.expectStatus {
				t.Fatalf("status mismatch: got %d want %d", tc.err.GetStatus(), tc.expectStatus)
			}
			if tcErr, ok := tc.err.(interface{ Error() string }); ok {
				if got := tcErr.Error(); got != tc.expectMessage {
					t.Fatalf("message mismatch: got %q want %q", got, tc.expectMessage)
				}
			} else {
				t.Fatalf("error does not implement Error()")
			}
		})
	}
}
