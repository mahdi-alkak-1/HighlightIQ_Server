# Deployment and CI/CD

## Overview
- Backend runs on a single EC2 Ubuntu instance.
- Go API runs on 127.0.0.1:8080 via systemd service `highlightiq-api`.
- Python clipper runs on 127.0.0.1:8090 via systemd service `highlightiq-clipper`.
- Nginx listens on port 80 and reverse-proxies to the API.
- MySQL runs on the same EC2 instance.

## Server layout
- Recordings: `/var/lib/highlightiq/recordings`
- Clips: `/var/lib/highlightiq/clips`
- Environment file: `highlightiq.env` (DB, JWT, paths)
- N8N webhook values are stored in `highlightiq.env`
- App directory (deploy target): `/home/ubuntu/HighlightIQ`

## Database
- Database: `highlightiq`
- User: `highlightiq`

## CI (GitHub Actions)
Workflow: `.github/workflows/ci.yml`
- Trigger: push to `main`
- Steps:
  - `go test ./...`
  - `go build -o highlightiq-api ./cmd/api/main.go`

## CD (GitHub Actions)
Workflow: `.github/workflows/cd.yml`
- Trigger: CI workflow completion (only on success)
- Deploy action: `appleboy/ssh-action@v1.0.0`
- Required secrets:
  - `EC2_HOST`
  - `EC2_USER`
  - `EC2_SSH_KEY`
- Server commands:
  - `cd /home/ubuntu/HighlightIQ`
  - `git pull`
  - `go mod download`
  - `go build -o highlightiq-api ./cmd/api/main.go`
  - `sudo systemctl restart highlightiq-api`
  - `sudo systemctl restart highlightiq-clipper`

## Deploy flow
1. Push to `main`.
2. CI runs tests and builds.
3. CD connects to EC2, pulls the repo, rebuilds, and restarts services.

## N8N backend trigger
When a clip export succeeds, the API can call an n8n webhook (if configured).

Required env vars in `highlightiq.env`:
- `N8N_PUBLISH_WEBHOOK_URL` (n8n webhook URL)
- `N8N_PUBLISH_WEBHOOK_AUTH` (value sent as `Authorization` header)
- `CLIPS_BASE_URL` (public base URL for clips, e.g. `http://51.44.15.133/clips`)
