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

type fakeAuthService struct{}

func (fakeAuthService) Register(ctx context.Context, in authsvc.RegisterInput) (authsvc.RegisterOutput, error) {
	return authsvc.RegisterOutput{
		User: authsvc.UserDTO{
			ID:    "test-uuid",
			Name:  in.Name,
			Email: in.Email,
		},
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}, nil
}

func TestAuthRegister(t *testing.T) {
	authHandler := authhandlers.New(fakeAuthService{})
	h := New(authHandler, nil, nil, nil)

	req := testutils.JSONRequest(http.MethodPost, "/auth/register", map[string]any{
		"name":     "Housam",
		"email":    "housam@test.com",
		"password": "password123",
	})

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusCreated, rr.Code, rr.Body.String())
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
