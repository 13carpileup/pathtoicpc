package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCodeforcesHandlersRequireParams(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name:    "user info requires handles",
			path:    "/api/user.info",
			handler: getUserInfo,
			wantErr: `missing required query parameter "handles"`,
		},
		{
			name:    "user status requires handle",
			path:    "/api/user.status",
			handler: getUserStatus,
			wantErr: `missing required query parameter "handle"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			tt.handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}

			var body errorResponse
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if body.Error != tt.wantErr {
				t.Fatalf("error = %q, want %q", body.Error, tt.wantErr)
			}
		})
	}
}
