// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"highlightiq-server/internal/config"
	"highlightiq-server/internal/db"
	authhandlers "highlightiq-server/internal/http/handlers/auth"
	clipcandhandlers "highlightiq-server/internal/http/handlers/clipcandidates"
	clipshandlers "highlightiq-server/internal/http/handlers/clips"
	recordinghandlers "highlightiq-server/internal/http/handlers/recordings"
	yphandlers "highlightiq-server/internal/http/handlers/youtubepublishes"
	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/router"

	"highlightiq-server/internal/integrations/clipper"
	"highlightiq-server/internal/integrations/n8n"

	clipcandidatesrepo "highlightiq-server/internal/repos/clipcandidates"
	clipsrepo "highlightiq-server/internal/repos/clips"
	recordingrepo "highlightiq-server/internal/repos/recordings"
	"highlightiq-server/internal/repos/users"
	youtubePublishesRepo "highlightiq-server/internal/repos/youtubepublishes"

	authsvc "highlightiq-server/internal/services/auth"
	clipcandidatessvc "highlightiq-server/internal/services/clipcandidates"
	clipssvc "highlightiq-server/internal/services/clips"
	recordingsvc "highlightiq-server/internal/services/recordings"
	ypsvc "highlightiq-server/internal/services/youtubepublishes"
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
	clipCandidatesRepo := clipcandidatesrepo.New(conn)
	clipsRepo := clipsrepo.New(conn)
	ypRepo := youtubePublishesRepo.New(conn)

	// services
	authService := authsvc.New(usersRepo, cfg.JWTSecret)
	recService := recordingsvc.New(recRepo, cfg.RecordingsDir)

	clipperClient := clipper.New("http://127.0.0.1:8090")
	clipCandidatesService := clipcandidatessvc.New(recRepo, clipCandidatesRepo, clipperClient)

	clipsDir := os.Getenv("CLIPS_DIR")
	if clipsDir == "" {
		clipsDir = "/var/lib/highlightiq/clips"
	}

	var publishNotifier clipssvc.PublishNotifier
	if cfg.N8NPublishWebhookURL != "" {
		publishNotifier = n8n.New(cfg.N8NPublishWebhookURL, cfg.N8NPublishWebhookAuth)
	}

	clipsService := clipssvc.New(clipsRepo, recRepo, clipsDir, cfg.ClipsBaseURL, publishNotifier)
	youtubePublishesService := ypsvc.New(clipsRepo, ypRepo)

	// handlers
	authHandler := authhandlers.New(authService)
	recHandler := recordinghandlers.New(recService)
	clipHandler := clipcandhandlers.New(clipCandidatesService)
	clipsHandler := clipshandlers.New(clipsService)
	youtubePublishesHandler := yphandlers.New(youtubePublishesService, cfg.N8NWebhookSecret)

	// middleware
	jwtAuth := middleware.NewJWTAuth(usersRepo, cfg.JWTSecret)

	// router
	r := router.New(authHandler, recHandler, clipHandler, clipsHandler, youtubePublishesHandler, jwtAuth.Middleware)

	log.Println("API listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
