package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeRegistration(t *testing.T) {
	tests := []struct {
		name    string
		req     registerRequest
		wantErr bool
	}{
		{
			name: "valid account",
			req: registerRequest{
				Email:    "USER@example.com",
				Username: "solver_123",
				Password: "correct horse",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			req: registerRequest{
				Email:    "not-email",
				Username: "solver_123",
				Password: "correct horse",
			},
			wantErr: true,
		},
		{
			name: "invalid username",
			req: registerRequest{
				Email:    "user@example.com",
				Username: "bad-name",
				Password: "correct horse",
			},
			wantErr: true,
		},
		{
			name: "short password",
			req: registerRequest{
				Email:    "user@example.com",
				Username: "solver_123",
				Password: "short",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, _, err := normalizeRegistration(tt.req)
			if tt.wantErr && err == nil {
				t.Fatal("expected an error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && email != "user@example.com" {
				t.Fatalf("email = %q, want %q", email, "user@example.com")
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	token, err := bearerToken(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "test-token" {
		t.Fatalf("token = %q, want %q", token, "test-token")
	}
}

func TestHashSessionToken(t *testing.T) {
	first := hashSessionToken("token")
	second := hashSessionToken("token")

	if first != second {
		t.Fatal("hashSessionToken should be deterministic")
	}
	if first == "token" {
		t.Fatal("hashSessionToken should not return the raw token")
	}
	if len(first) != 64 {
		t.Fatalf("hash length = %d, want 64", len(first))
	}
}
