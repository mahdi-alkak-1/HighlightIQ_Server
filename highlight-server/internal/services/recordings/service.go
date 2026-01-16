package recordings

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	recRepo "highlightiq-server/internal/repos/recordings"
)

type Service struct {
	repo     *recRepo.Repo
	baseDir  string
	maxBytes int64
}

func New(repo *recRepo.Repo, baseDir string) *Service {
	if baseDir == "" {
		baseDir = "./storage/recordings"
	}
	return &Service{
		repo:     repo,
		baseDir:  baseDir,
		maxBytes: 1_000_000_000, // 1GB
	}
}

func (s *Service) Create(ctx context.Context, userID int64, title string, originalName string, fileBytes []byte) (recRepo.Recording, error) {
	recUUID := uuid.NewString()

	if title == "" {
		title = filenameNoExt(originalName)
	}

	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return recRepo.Recording{}, err
	}

	fileName := recUUID + "_" + sanitizeFileName(originalName)
	fullPath := filepath.Join(s.baseDir, fileName)

	if err := os.WriteFile(fullPath, fileBytes, 0o644); err != nil {
		return recRepo.Recording{}, err
	}

	rec, err := s.repo.Create(ctx, recRepo.CreateParams{
		UUID:            recUUID,
		UserID:          userID,
		Title:           title,
		OriginalName:    originalName,
		StoragePath:     fullPath,
		DurationSeconds: 0,
		Status:          "uploaded",
	})
	if err != nil {
		// If DB insert fails, clean up the saved file
		_ = os.Remove(fullPath)
		return recRepo.Recording{}, err
	}

	return rec, nil
}

func (s *Service) List(ctx context.Context, userID int64) ([]recRepo.Recording, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID int64, recUUID string) (recRepo.Recording, error) {
	return s.repo.GetByUUIDForUser(ctx, userID, recUUID, 0)
}

func (s *Service) UpdateTitle(ctx context.Context, userID int64, recUUID string, title string) error {
	if title == "" {
		return recRepo.ErrNotFound // will be mapped to 422 by handler; keep it simple
	}
	return s.repo.UpdateTitleByUUIDForUser(ctx, userID, recUUID, title)
}

func (s *Service) Delete(ctx context.Context, userID int64, recUUID string) error {
	path, err := s.repo.DeleteByUUIDForUser(ctx, userID, recUUID)
	if err != nil {
		return err
	}
	// Best-effort file delete (if it fails, DB row is already gone)
	_ = os.Remove(path)
	return nil
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	// remove dangerous path chars
	name = strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return -1
		default:
			return r
		}
	}, name)
	if name == "" {
		return "upload.mp4"
	}
	return name
}

func filenameNoExt(name string) string {
	base := filepath.Base(name)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// Optional helper if you want to format time later
func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
