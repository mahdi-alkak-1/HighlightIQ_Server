// cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"highlightiq-server/internal/config"
	"highlightiq-server/internal/db"

	authhandlers "highlightiq-server/internal/http/handlers/auth"
	clipcandhandlers "highlightiq-server/internal/http/handlers/clipcandidates"
	clipshandlers "highlightiq-server/internal/http/handlers/clips"
	recordinghandlers "highlightiq-server/internal/http/handlers/recordings"
	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/router"

	"highlightiq-server/internal/integrations/clipper"

	clipcandidatesrepo "highlightiq-server/internal/repos/clipcandidates"
	clipsrepo "highlightiq-server/internal/repos/clips"
	recordingrepo "highlightiq-server/internal/repos/recordings"
	"highlightiq-server/internal/repos/users"

	authsvc "highlightiq-server/internal/services/auth"
	clipcandidatessvc "highlightiq-server/internal/services/clipcandidates"
	clipssvc "highlightiq-server/internal/services/clips"
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
	clipCandidatesRepo := clipcandidatesrepo.New(conn)
	clipsRepo := clipsrepo.New(conn)

	// services
	authService := authsvc.New(usersRepo, cfg.JWTSecret)
	recService := recordingsvc.New(recRepo, cfg.RecordingsDir)

	clipperClient := clipper.New("http://127.0.0.1:8090")
	clipCandidatesService := clipcandidatessvc.New(recRepo, clipCandidatesRepo, clipperClient)

	// Store clips in D:\clips
	clipsService := clipssvc.New(clipsRepo, recRepo, `D:\clips`)

	// handlers
	authHandler := authhandlers.New(authService)
	recHandler := recordinghandlers.New(recService)
	clipHandler := clipcandhandlers.New(clipCandidatesService)
	clipsHandler := clipshandlers.New(clipsService)

	// middleware
	jwtAuth := middleware.NewJWTAuth(usersRepo, cfg.JWTSecret)

	// router
	r := router.New(authHandler, recHandler, clipHandler, clipsHandler, jwtAuth.Middleware)

	log.Println("API listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
