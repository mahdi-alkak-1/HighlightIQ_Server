package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	recordinghandlers "highlightiq-server/internal/http/handlers/recordings"
	"highlightiq-server/internal/http/middleware"
	recRepo "highlightiq-server/internal/repos/recordings"
	"highlightiq-server/internal/testutils"
)

type fakeRecordingsService struct{}

func (fakeRecordingsService) Create(ctx context.Context, userID int64, title string, originalName string, fileBytes []byte) (recRepo.Recording, error) {
	return recRepo.Recording{
		ID:           1,
		UUID:         "rec-uuid-1",
		UserID:       userID,
		Title:        title,
		OriginalName: originalName,
		StoragePath:  "D:\\recordings\\rec-uuid-1_test.mp4",
		Status:       "uploaded",
	}, nil
}

func (fakeRecordingsService) List(ctx context.Context, userID int64) ([]recRepo.Recording, error) {
	return []recRepo.Recording{
		{
			ID:           1,
			UUID:         "rec-uuid-1",
			UserID:       userID,
			Title:        "fortnite",
			OriginalName: "fortnite.mp4",
			StoragePath:  "D:\\recordings\\rec-uuid-1_fortnite.mp4",
			Status:       "uploaded",
		},
	}, nil
}

func (fakeRecordingsService) Get(ctx context.Context, userID int64, recUUID string) (recRepo.Recording, error) {
	return recRepo.Recording{
		ID:           1,
		UUID:         recUUID,
		UserID:       userID,
		Title:        "fortnite",
		OriginalName: "fortnite.mp4",
		StoragePath:  "D:\\recordings\\rec-uuid-1_fortnite.mp4",
		Status:       "uploaded",
	}, nil
}

func (fakeRecordingsService) UpdateTitle(ctx context.Context, userID int64, recUUID string, title string) error {
	return nil
}

func (fakeRecordingsService) Delete(ctx context.Context, userID int64, recUUID string) error {
	return nil
}

func fakeAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithAuthUser(r.Context(), middleware.AuthUser{
			ID:    1,
			UUID:  "user-uuid-1",
			Email: "user@test.com",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TestRecordingsList(t *testing.T) {
	recHandler := recordinghandlers.New(fakeRecordingsService{})
	h := New(nil, recHandler, nil, fakeAuthMW)

	req := httptest.NewRequest(http.MethodGet, "/recordings", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if resp["data"] == nil {
		t.Fatalf("expected data in response")
	}
}

func TestRecordingsUpdateTitle(t *testing.T) {
	recHandler := recordinghandlers.New(fakeRecordingsService{})
	h := New(nil, recHandler, nil, fakeAuthMW)

	req := testutils.JSONRequest(http.MethodPatch, "/recordings/rec-uuid-1", map[string]any{
		"title": "new title",
	})

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}
