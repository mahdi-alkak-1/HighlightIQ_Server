package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"highlightiq-server/internal/testutils"
)

func TestAuthRegister(t *testing.T) {
	h := New()

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
