package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authhandlers "highlightiq-server/internal/http/handlers/auth"
	authsvc "highlightiq-server/internal/services/auth"
	"highlightiq-server/internal/testutils"
)

func (fakeAuthService) Login(ctx context.Context, in authsvc.LoginInput) (authsvc.RegisterOutput, error) {
	return authsvc.RegisterOutput{
		User: authsvc.UserDTO{
			ID:    "test-uuid",
			Name:  "Test User",
			Email: in.Email,
		},
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}, nil
}

func TestAuthLogin(t *testing.T) {
	authHandler := authhandlers.New(fakeAuthService{})
	h := New(authHandler, nil, nil, nil)

	req := testutils.JSONRequest(http.MethodPost, "/auth/login", map[string]any{
		"email":    "housam@test.com",
		"password": "password123",
	})

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if resp["access_token"] == nil {
		t.Fatalf("expected access_token in response")
	}
	if resp["user"] == nil {
		t.Fatalf("expected user in response")
	}
}
