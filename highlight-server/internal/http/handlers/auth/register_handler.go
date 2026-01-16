package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	resp "highlightiq-server/internal/http/response"
	authreq "highlightiq-server/internal/requests/auth"
	authsvc "highlightiq-server/internal/services/auth"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authreq.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.JSON(w, http.StatusBadRequest, map[string]any{
			"message": "invalid JSON payload",
		})
		return
	}

	if err := req.Validate(); err != nil {
		// If it's our structured validation error, return it as a map
		if verr, ok := err.(authreq.ValidationError); ok {
			resp.JSON(w, http.StatusUnprocessableEntity, map[string]any{
				"message": "validation error",
				"errors":  verr,
			})
			return
		}

		resp.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"message": "validation error",
			"errors":  err.Error(),
		})
		return
	}

	out, err := h.svc.Register(r.Context(), authsvc.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, authsvc.ErrEmailTaken) {
			resp.JSON(w, http.StatusConflict, map[string]any{
				"message": "email already registered",
			})
			return
		}

		resp.JSON(w, http.StatusInternalServerError, map[string]any{
			"message": "internal server error",
		})
		return
	}

	resp.JSON(w, http.StatusCreated, out)
}
