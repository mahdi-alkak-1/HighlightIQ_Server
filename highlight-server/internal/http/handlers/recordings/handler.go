package recordings

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/response"
	recRepo "highlightiq-server/internal/repos/recordings"
	recReq "highlightiq-server/internal/requests/recordings"
)

type RecordingService interface {
	Create(ctx context.Context, userID int64, title string, originalName string, fileBytes []byte) (recRepo.Recording, error)
	List(ctx context.Context, userID int64) ([]recRepo.Recording, error)
	Get(ctx context.Context, userID int64, recUUID string) (recRepo.Recording, error)
	UpdateTitle(ctx context.Context, userID int64, recUUID string, title string) error
	Delete(ctx context.Context, userID int64, recUUID string) error
}

type Handler struct {
	svc RecordingService
}

func New(svc RecordingService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
		return
	}

	// 1GB hard limit
	r.Body = http.MaxBytesReader(w, r.Body, 1_000_000_000)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"message": "validation error",
			"errors":  map[string]string{"file": "file is required"},
		})
		return
	}
	defer file.Close()

	title := strings.TrimSpace(r.FormValue("title"))

	b, err := io.ReadAll(file)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "failed to read file"})
		return
	}

	rec, err := h.svc.Create(r.Context(), u.ID, title, header.Filename, b)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "internal server error"})
		return
	}

	response.JSON(w, http.StatusCreated, rec)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
		return
	}

	recs, err := h.svc.List(r.Context(), u.ID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "internal server error"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"data": recs})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
		return
	}

	recUUID := chi.URLParam(r, "uuid")
	rec, err := h.svc.Get(r.Context(), u.ID, recUUID)
	if err != nil {
		if errors.Is(err, recRepo.ErrNotFound) {
			response.JSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "internal server error"})
		return
	}

	response.JSON(w, http.StatusOK, rec)
}

func (h *Handler) UpdateTitle(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
		return
	}

	recUUID := chi.URLParam(r, "uuid")

	var req recReq.UpdateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		if verr, ok := err.(recReq.ValidationError); ok {
			response.JSON(w, http.StatusUnprocessableEntity, map[string]any{
				"message": "validation error",
				"errors":  verr,
			})
			return
		}
		response.JSON(w, http.StatusUnprocessableEntity, map[string]any{"message": "validation error"})
		return
	}

	if err := h.svc.UpdateTitle(r.Context(), u.ID, recUUID, req.Title); err != nil {
		if errors.Is(err, recRepo.ErrNotFound) {
			response.JSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "internal server error"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"message": "ok"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
		return
	}

	recUUID := chi.URLParam(r, "uuid")

	if err := h.svc.Delete(r.Context(), u.ID, recUUID); err != nil {
		if errors.Is(err, recRepo.ErrNotFound) {
			response.JSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "internal server error"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"message": "deleted"})
}
