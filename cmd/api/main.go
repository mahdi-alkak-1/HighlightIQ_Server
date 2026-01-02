package main

import (
	"log"
	"net/http"

	"highlightiq-server/internal/config"
	"highlightiq-server/internal/db"

	authhandlers "highlightiq-server/internal/http/handlers/auth"
	clipcandhandlers "highlightiq-server/internal/http/handlers/clipcandidates"
	recordinghandlers "highlightiq-server/internal/http/handlers/recordings"
	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/router"

	"highlightiq-server/internal/integrations/clipper"

	clipcandidatesrepo "highlightiq-server/internal/repos/clipcandidates"
	recordingrepo "highlightiq-server/internal/repos/recordings"
	"highlightiq-server/internal/repos/users"

	authsvc "highlightiq-server/internal/services/auth"
	clipcandidatessvc "highlightiq-server/internal/services/clipcandidates"
	recordingsvc "highlightiq-server/internal/services/recordings"
)

func main() {
	cfg := config.Load()

	conn, err := db.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer conn.Close()

	// repos
	usersRepo := users.New(conn)
	recRepo := recordingrepo.New(conn)
	clipRepo := clipcandidatesrepo.New(conn)

	// services
	authService := authsvc.New(usersRepo, cfg.JWTSecret)
	recService := recordingsvc.New(recRepo, cfg.RecordingsDir)

	// IMPORTANT: this MUST point to your python FastAPI service
	// that exposes POST http://127.0.0.1:8090/detect-candidates
	clipperClient := clipper.New("http://127.0.0.1:8090")

	clipService := clipcandidatessvc.New(recRepo, clipRepo, clipperClient)

	// handlers
	authHandler := authhandlers.New(authService)
	recHandler := recordinghandlers.New(recService)
	clipHandler := clipcandhandlers.New(clipService)

	// middleware
	jwtAuth := middleware.NewJWTAuth(usersRepo, cfg.JWTSecret)

	// router (NOW clip handler is wired)
	r := router.New(authHandler, recHandler, clipHandler, jwtAuth.Middleware)

	log.Println("API listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
