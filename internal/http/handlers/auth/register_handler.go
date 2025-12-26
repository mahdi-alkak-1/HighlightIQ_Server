package auth

import (
	"encoding/json"
	"net/http"

	"highlightiq-server/internal/http/response"
	authreq "highlightiq-server/internal/requests/auth"
	authsvc "highlightiq-server/internal/services/auth"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var req authreq.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := authreq.Validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "validation failed")
		return
	}

	result, status, errMsg := authsvc.Register(req.Name, req.Email)
	if errMsg != "" {
		response.Error(w, status, errMsg)
		return
	}

	response.JSON(w, status, result)
}
