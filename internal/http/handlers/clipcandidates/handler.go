package clipcandidates

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/response"
	reqs "highlightiq-server/internal/requests/clipcandidates"
	svc "highlightiq-server/internal/services/clipcandidates"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler {
	return &Handler{svc: s}
}

// POST /recordings/{uuid}/clip-candidates/detect
func (h *Handler) Detect(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
		return
	}

	recordingUUID := chi.URLParam(r, "uuid")

	var req reqs.DetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "validation failed"})
		return
	}

	inserted, err := h.svc.DetectAndStore(r.Context(), u.ID, svc.DetectInput{
		RecordingUUID:     recordingUUID,
		ClipLengthSeconds: req.ClipLengthSeconds,
		Threshold:         req.Threshold,
		MinClipSeconds:    req.MinClipSeconds,
		MaxCandidates:     req.MaxCandidates,
		MinSpacingSeconds: req.MinSpacingSeconds,
	})
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, map[string]any{"message": "recording not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "failed to detect candidates"})
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"inserted": inserted,
	})
}

// GET /recordings/{uuid}/clip-candidates
func (h *Handler) ListByRecording(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
		return
	}

	recordingUUID := chi.URLParam(r, "uuid")

	items, err := h.svc.ListByRecordingUUID(r.Context(), u.ID, recordingUUID)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, map[string]any{"message": "recording not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "failed to list candidates"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// PATCH /clip-candidates/{id}
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "invalid id"})
		return
	}

	var req reqs.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "validation failed"})
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "failed to update status"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /clip-candidates/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{"message": "invalid id"})
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "failed to delete candidate"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
