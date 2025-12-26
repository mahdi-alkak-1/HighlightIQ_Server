package router

import (
	"net/http"

	auth "highlightiq-server/internal/http/handlers/auth"

	"github.com/go-chi/chi/v5"
)

func New() http.Handler {
	r := chi.NewRouter()

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Auth routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
	})

	return r
}
