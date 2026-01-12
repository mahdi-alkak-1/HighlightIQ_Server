package router

import (
	"net/http"

	authhandlers "highlightiq-server/internal/http/handlers/auth"
	clipcandhandlers "highlightiq-server/internal/http/handlers/clipcandidates"
	clipshandlers "highlightiq-server/internal/http/handlers/clips"
	recordinghandlers "highlightiq-server/internal/http/handlers/recordings"

	"github.com/go-chi/chi/v5"
)

func New(
	authHandler *authhandlers.Handler,
	recordingsHandler *recordinghandlers.Handler,
	clipCandidatesHandler *clipcandhandlers.Handler,
	clipsHandler *clipshandlers.Handler,
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Public auth routes
	if authHandler != nil {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
	}

	// Protected routes (JWT required)
	if authMiddleware != nil {
		r.Group(func(pr chi.Router) {
			pr.Use(authMiddleware)

			// Recordings CRUD
			if recordingsHandler != nil {
				pr.Route("/recordings", func(rr chi.Router) {
					rr.Post("/", recordingsHandler.Create)
					rr.Get("/", recordingsHandler.List)

					rr.Route("/{uuid}", func(r3 chi.Router) {
						r3.Get("/", recordingsHandler.Get)
						r3.Patch("/", recordingsHandler.UpdateTitle)
						r3.Delete("/", recordingsHandler.Delete)

						// Nested clip candidates for a recording
						if clipCandidatesHandler != nil {
							r3.Route("/clip-candidates", func(cr chi.Router) {
								cr.Get("/", clipCandidatesHandler.ListByRecording)
								cr.Post("/detect", clipCandidatesHandler.Detect)
							})
						}
					})
				})
			}

			// Candidate actions by id
			if clipCandidatesHandler != nil {
				pr.Route("/clip-candidates/{id}", func(cr chi.Router) {
					cr.Patch("/", clipCandidatesHandler.UpdateStatus)
					cr.Delete("/", clipCandidatesHandler.Delete)
				})
			}

			// Clips CRUD + export
			if clipsHandler != nil {
				pr.Route("/clips", func(cr chi.Router) {
					cr.Post("/", clipsHandler.Create)
					cr.Get("/", clipsHandler.List)

					cr.Route("/{id}", func(r3 chi.Router) {
						r3.Get("/", clipsHandler.Get)
						r3.Patch("/", clipsHandler.Update)
						r3.Delete("/", clipsHandler.Delete)
						r3.Post("/export", clipsHandler.Export)
					})
				})
			}
		})
	}

	return r
}
