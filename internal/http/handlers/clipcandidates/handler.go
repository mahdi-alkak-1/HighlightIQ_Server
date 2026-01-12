package clipcandidates

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/response"
	reqs "highlightiq-server/internal/requests/clipcandidates"
	svc "highlightiq-server/internal/services/clipcandidates"
	"io"
	"log"
	"net/http"
	"strconv"
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
		response.JSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
		return
	}

	recordingUUID := chi.URLParam(r, "uuid")

	var req reqs.DetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		response.JSON(w, http.StatusBadRequest, map[string]string{"message": "invalid JSON payload"})
		return
	}

	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"message": "validation failed"})
		return
	}

	inserted, err := h.svc.DetectAndStore(r.Context(), u.ID, svc.DetectInput{
		RecordingUUID:      recordingUUID,
		MaxClipSeconds:     req.MaxClipSeconds,
		PreRollSeconds:     req.PreRollSeconds,
		PostRollSeconds:    req.PostRollSeconds,
		MinClipSeconds:     req.MinClipSeconds,
		SampleFPS:          req.SampleFPS,
		MinSpacingSeconds:  req.MinSpacingSeconds,
		MergeGapSeconds:    req.MergeGapSeconds,
		ElimMatchThreshold: req.ElimMatchThreshold,
		MinConsecutiveHits: req.MinConsecutiveHits,
		CooldownSeconds:    req.CooldownSeconds,
	})

	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, map[string]string{"message": "recording not found"})
			return
		}
		log.Printf("DetectAndStore failed: %v", err)
		response.JSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to detect candidates"})
		return
	}

	response.JSON(w, http.StatusCreated, map[string]int64{
		"inserted": inserted,
	})
}

// GET /recordings/{uuid}/clip-candidates
func (h *Handler) ListByRecording(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
		return
	}

	recordingUUID := chi.URLParam(r, "uuid")

	items, err := h.svc.ListByRecordingUUID(r.Context(), u.ID, recordingUUID)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, map[string]string{"message": "recording not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to list candidates"})
		return
	}

	// strict type response (no any)
	type resp struct {
		Items interface{} `json:"items"`
	}
	response.JSON(w, http.StatusOK, resp{Items: items})
}

// PATCH /clip-candidates/{id}
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"message": "invalid id"})
		return
	}

	var req reqs.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"message": "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"message": "validation failed"})
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to update status"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /clip-candidates/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"message": "invalid id"})
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to delete candidate"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
