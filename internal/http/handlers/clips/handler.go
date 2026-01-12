package clips

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/response"
	reqs "highlightiq-server/internal/requests/clips"
	svc "highlightiq-server/internal/services/clips"
	"log"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler {
	return &Handler{svc: s}
}

type messageResponse struct {
	Message string `json:"message"`
}

type listResponse struct {
	Data []interface{} `json:"-"`
}

type clipsListResponse struct {
	Data []struct{} `json:"-"`
}

// POST /clips
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	var req reqs.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "validation failed"})
		return
	}
	if req.EndMS <= req.StartMS {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "end_ms must be > start_ms"})
		return
	}

	clip, err := h.svc.Create(r.Context(), u.ID, svc.CreateInput{
		RecordingUUID: req.RecordingUUID,
		CandidateID:   req.CandidateID,
		Title:         req.Title,
		Caption:       req.Caption,
		StartMS:       req.StartMS,
		EndMS:         req.EndMS,
	})
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "recording not found"})
			return
		}
		if err == svc.ErrBadInput {
			response.JSON(w, http.StatusBadRequest, messageResponse{Message: "bad input"})
			return
		}

		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to create clip"})
		return
	}

	response.JSON(w, http.StatusCreated, clip)
}

// GET /clips?recording_uuid=...
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	recUUID := r.URL.Query().Get("recording_uuid")
	var recPtr *string
	if recUUID != "" {
		recPtr = &recUUID
	}

	items, err := h.svc.List(r.Context(), u.ID, recPtr)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "recording not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to list clips"})
		return
	}

	type resp struct {
		Data interface{} `json:"data"`
	}
	response.JSON(w, http.StatusOK, resp{Data: items})
}

// GET /clips/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid id"})
		return
	}

	clip, err := h.svc.Get(r.Context(), u.ID, id)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to get clip"})
		return
	}

	response.JSON(w, http.StatusOK, clip)
}

// PATCH /clips/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid id"})
		return
	}

	var req reqs.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "validation failed"})
		return
	}

	clip, err := h.svc.Update(r.Context(), u.ID, id, svc.UpdateInput{
		Title:   req.Title,
		Caption: req.Caption,
		StartMS: req.StartMS,
		EndMS:   req.EndMS,
		Status:  req.Status,
	})
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		if err == svc.ErrBadInput {
			response.JSON(w, http.StatusBadRequest, messageResponse{Message: "bad input"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to update clip"})
		return
	}

	response.JSON(w, http.StatusOK, clip)
}

// DELETE /clips/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid id"})
		return
	}

	if err := h.svc.Delete(r.Context(), u.ID, id); err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to delete clip"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /clips/{id}/export
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid id"})
		return
	}

	clip, err := h.svc.Export(r.Context(), u.ID, id)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		log.Printf("Create clip failed: %v", err)
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to export clip"})
		return
	}

	response.JSON(w, http.StatusOK, clip)
}
