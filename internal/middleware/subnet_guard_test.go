package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNewSubnetGuard(t *testing.T) {
	tests := []struct {
		name       string
		subNetMask string
		wantErr    bool
		wantNilNet bool
	}{
		{
			name:       "valid IPv4 CIDR",
			subNetMask: "192.168.1.0/24",
			wantErr:    false,
			wantNilNet: false,
		},
		{
			name:       "valid IPv6 CIDR",
			subNetMask: "2001:db8::/32",
			wantErr:    false,
			wantNilNet: false,
		},
		{
			name:       "empty mask disables check",
			subNetMask: "",
			wantErr:    false,
			wantNilNet: true,
		},
		{
			name:       "invalid mask without prefix",
			subNetMask: "192.168.1.1",
			wantErr:    true,
		},
		{
			name:       "invalid prefix length",
			subNetMask: "192.168.1.0/33",
			wantErr:    true,
		},
		{
			name:       "garbage string",
			subNetMask: "not-a-cidr",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard, err := NewSubnetGuard(tt.subNetMask)

			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSubnetGuard(%q) error = %v, wantErr %v", tt.subNetMask, err, tt.wantErr)
			}

			if !tt.wantErr {
				if guard == nil {
					t.Fatal("expected non-nil guard")
				}
				if (guard.trustedNet == nil) != tt.wantNilNet {
					t.Errorf("trustedNet == nil is %v, want %v", guard.trustedNet == nil, tt.wantNilNet)
				}
			}
		})
	}
}

func TestSubnetGuard_Middleware(t *testing.T) {
	tests := []struct {
		name           string
		subNetMask     string
		xRealIPHeader  string
		wantStatusCode int
		wantNextCalled bool
	}{
		{
			name:           "allowed IP inside subnet",
			subNetMask:     "192.168.1.0/24",
			xRealIPHeader:  "192.168.1.42",
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "forbidden IP outside subnet",
			subNetMask:     "192.168.1.0/24",
			xRealIPHeader:  "10.0.0.5",
			wantStatusCode: http.StatusForbidden,
			wantNextCalled: false,
		},
		{
			name:           "forbidden when X-Real-IP header is missing",
			subNetMask:     "192.168.1.0/24",
			xRealIPHeader:  "",
			wantStatusCode: http.StatusForbidden,
			wantNextCalled: false,
		},
		{
			name:           "forbidden when X-Real-IP is malformed",
			subNetMask:     "192.168.1.0/24",
			xRealIPHeader:  "invalid-ip",
			wantStatusCode: http.StatusForbidden,
			wantNextCalled: false,
		},
		{
			name:           "allowed any request when subnet mask is empty",
			subNetMask:     "",
			xRealIPHeader:  "",
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "allowed IPv6 inside subnet",
			subNetMask:     "2001:db8::/32",
			xRealIPHeader:  "2001:db8::1",
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard, err := NewSubnetGuard(tt.subNetMask)
			if err != nil {
				t.Fatalf("unexpected error creating guard: %v", err)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/update/", nil)
			if tt.xRealIPHeader != "" {
				req.Header.Set("X-Real-IP", tt.xRealIPHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalled := false
			nextHandler := func(c echo.Context) error {
				nextCalled = true
				return c.NoContent(http.StatusOK)
			}

			mw := guard.Middleware()
			handler := mw(nextHandler)

			if err := handler(c); err != nil {
				t.Fatalf("handler returned unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}

			if nextCalled != tt.wantNextCalled {
				t.Errorf("nextCalled = %v, want %v", nextCalled, tt.wantNextCalled)
			}
		})
	}
}
