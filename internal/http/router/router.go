package router

import (
	"net/http"

	authhandlers "highlightiq-server/internal/http/handlers/auth"

	"github.com/go-chi/chi/v5"
)

func New(authHandler *authhandlers.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	if authHandler != nil {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
	}

	return r
}
