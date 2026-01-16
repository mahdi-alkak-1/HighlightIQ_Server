package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	resp "highlightiq-server/internal/http/response"
	authreq "highlightiq-server/internal/requests/auth"
	authsvc "highlightiq-server/internal/services/auth"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authreq.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.JSON(w, http.StatusBadRequest, map[string]any{
			"message": "invalid JSON payload",
		})
		return
	}

	if err := req.Validate(); err != nil {
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

	out, err := h.svc.Login(r.Context(), authsvc.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidCredentials) {
			resp.JSON(w, http.StatusUnauthorized, map[string]any{
				"message": "invalid credentials",
			})
			return
		}

		resp.JSON(w, http.StatusInternalServerError, map[string]any{
			"message": "internal server error",
		})
		return
	}

	resp.JSON(w, http.StatusOK, out)
}
