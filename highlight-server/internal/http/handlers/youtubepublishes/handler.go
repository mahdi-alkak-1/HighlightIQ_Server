package youtubepublishes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/response"
	reqs "highlightiq-server/internal/requests/youtubepublishes"
	svc "highlightiq-server/internal/services/youtubepublishes"
)

type Handler struct {
	svc    *svc.Service
	secret string
}

func New(s *svc.Service, secret string) *Handler {
	return &Handler{svc: s, secret: secret}
}

type messageResponse struct {
	Message string `json:"message"`
}

// POST /clips/{id}/youtube-publishes
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	clipID, err := parseIDParam(r, "id")
	if err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid clip id"})
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

	status := ""
	if req.Status != nil {
		status = *req.Status
	}

	created, err := h.svc.Create(r.Context(), u.ID, clipID, svc.CreateInput{
		YoutubeVideoID: req.YoutubeVideoID,
		YoutubeURL:     req.YoutubeURL,
		Status:         status,
		PublishedAt:    req.PublishedAt,
		LastSyncedAt:   req.LastSyncedAt,
		Views:          0,
		Likes:          0,
		Comments:       0,
		Analytics:      nil,
	})
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "clip not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to create youtube publish"})
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

// GET /clips/{id}/youtube-publishes
func (h *Handler) ListByClip(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	clipID, err := parseIDParam(r, "id")
	if err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid clip id"})
		return
	}

	items, err := h.svc.ListByClip(r.Context(), u.ID, clipID)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "clip not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to list youtube publishes"})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"data": items})
}

// PATCH /youtube-publishes/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	id, err := parseIDParam(r, "id")
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

	updated, err := h.svc.Update(r.Context(), u.ID, id, svc.UpdateInput{
		YoutubeURL:   req.YoutubeURL,
		Status:       req.Status,
		PublishedAt:  req.PublishedAt,
		LastSyncedAt: req.LastSyncedAt,
	})
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to update youtube publish"})
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

// POST /internal/youtube-publishes
func (h *Handler) InternalCreate(w http.ResponseWriter, r *http.Request) {
	if !h.checkSecret(r) {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	var req reqs.InternalCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "validation failed"})
		return
	}

	status := ""
	if req.Status != nil {
		status = *req.Status
	}

	created, err := h.svc.CreateInternal(r.Context(), req.ClipID, svc.CreateInput{
		YoutubeVideoID: req.YoutubeVideoID,
		YoutubeURL:     req.YoutubeURL,
		Status:         status,
		PublishedAt:    req.PublishedAt,
		LastSyncedAt:   req.LastSyncedAt,
		Views:          req.Views,
		Likes:          req.Likes,
		Comments:       req.Comments,
		Analytics:      normalizeAnalytics(req.Analytics),
	})
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "clip not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to create youtube publish"})
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

// GET /internal/youtube-publishes
func (h *Handler) InternalList(w http.ResponseWriter, r *http.Request) {
	if !h.checkSecret(r) {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	ids, err := h.svc.ListVideoIDs(r.Context())
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to list youtube video ids"})
		return
	}

	type item struct {
		YoutubeVideoID string `json:"youtube_video_id"`
	}
	out := make([]item, 0, len(ids))
	for _, id := range ids {
		out = append(out, item{YoutubeVideoID: id})
	}

	response.JSON(w, http.StatusOK, map[string]any{"data": out})
}

// GET /internal/youtube-publishes/{youtube_video_id}
func (h *Handler) InternalGetByVideoID(w http.ResponseWriter, r *http.Request) {
	if !h.checkSecret(r) {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	videoID := chi.URLParam(r, "youtube_video_id")
	if videoID == "" {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid youtube video id"})
		return
	}

	item, err := h.svc.GetByVideoID(r.Context(), videoID)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to get youtube publish"})
		return
	}

	response.JSON(w, http.StatusOK, item)
}

// POST /internal/youtube-publishes/mark-deleted
func (h *Handler) InternalMarkDeleted(w http.ResponseWriter, r *http.Request) {
	if !h.checkSecret(r) {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	var req reqs.InternalMarkDeletedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "validation failed"})
		return
	}

	ts := req.LastSyncedAt
	if ts == nil {
		now := time.Now().UTC()
		ts = &now
	}

	updated, err := h.svc.MarkDeletedByVideoID(r.Context(), req.YoutubeVideoID, ts)
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to mark youtube publish deleted"})
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

// POST /internal/youtube-publishes/metrics
func (h *Handler) InternalUpdateMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.checkSecret(r) {
		response.JSON(w, http.StatusUnauthorized, messageResponse{Message: "unauthorized"})
		return
	}

	var req reqs.InternalMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "invalid JSON payload"})
		return
	}
	if err := req.Validate(); err != nil {
		response.JSON(w, http.StatusBadRequest, messageResponse{Message: "validation failed"})
		return
	}

	updated, err := h.svc.UpdateByVideoID(r.Context(), req.YoutubeVideoID, svc.UpdateInput{
		Views:        req.Views,
		Likes:        req.Likes,
		Comments:     req.Comments,
		PublishedAt:  req.PublishedAt,
		LastSyncedAt: req.LastSyncedAt,
		Analytics:    normalizeAnalytics(req.Analytics),
	})
	if err != nil {
		if err == svc.ErrNotFound {
			response.JSON(w, http.StatusNotFound, messageResponse{Message: "not found"})
			return
		}
		response.JSON(w, http.StatusInternalServerError, messageResponse{Message: "failed to update youtube publish"})
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func parseIDParam(r *http.Request, param string) (int64, error) {
	idStr := chi.URLParam(r, param)
	return strconv.ParseInt(idStr, 10, 64)
}

func normalizeAnalytics(raw *json.RawMessage) *string {
	if raw == nil {
		return nil
	}

	b := bytes.TrimSpace(*raw)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}

	s := string(b)
	return &s
}

func (h *Handler) checkSecret(r *http.Request) bool {
	if h.secret == "" {
		return false
	}
	return r.Header.Get("X-N8N-SECRET") == h.secret
}
