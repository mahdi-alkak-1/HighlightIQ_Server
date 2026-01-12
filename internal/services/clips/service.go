// internal/services/clips/service.go
package clips

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	clipsrepo "highlightiq-server/internal/repos/clips"
	recordingsrepo "highlightiq-server/internal/repos/recordings"
)

var ErrNotFound = errors.New("clips: not found")
var ErrBadInput = errors.New("clips: bad input")

type Service struct {
	clipsRepo      *clipsrepo.Repo
	recordingsRepo *recordingsrepo.Repo
	clipsDir       string
	ffmpegPath     string
}

func New(clipsRepo *clipsrepo.Repo, recordingsRepo *recordingsrepo.Repo, clipsDir string) *Service {
	return &Service{
		clipsRepo:      clipsRepo,
		recordingsRepo: recordingsRepo,
		clipsDir:       clipsDir,
		ffmpegPath:     resolveFFmpegPath(),
	}
}

func resolveFFmpegPath() string {
	// 1) Allow explicit override (recommended on Windows)
	// Can be either:
	//   - full path to ffmpeg.exe
	//   - directory that contains ffmpeg.exe
	if v := strings.TrimSpace(os.Getenv("FFMPEG_PATH")); v != "" {
		// If it's a directory, append ffmpeg(.exe)
		if st, err := os.Stat(v); err == nil && st.IsDir() {
			if runtime.GOOS == "windows" {
				return filepath.Join(v, "ffmpeg.exe")
			}
			return filepath.Join(v, "ffmpeg")
		}
		return v
	}

	// 2) Try PATH lookup
	// On Windows, try both "ffmpeg" and "ffmpeg.exe"
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("ffmpeg.exe"); err == nil {
			return p
		}
	}

	// 3) Fallback (will error at runtime with a clear message)
	return "ffmpeg"
}

type CreateInput struct {
	RecordingUUID string
	CandidateID   *int64

	Title   string
	Caption *string

	StartMS int
	EndMS   int
}

func (s *Service) Create(ctx context.Context, userID int64, in CreateInput) (clipsrepo.Clip, error) {
	if in.EndMS <= in.StartMS {
		return clipsrepo.Clip{}, ErrBadInput
	}

	rec, err := s.recordingsRepo.GetByUUIDForUser(ctx, userID, in.RecordingUUID, 0)
	if err != nil {
		return clipsrepo.Clip{}, ErrNotFound
	}

	return s.clipsRepo.Create(ctx, clipsrepo.CreateParams{
		UserID:      userID,
		RecordingID: rec.ID,
		CandidateID: in.CandidateID,
		Title:       in.Title,
		Caption:     in.Caption,
		StartMS:     in.StartMS,
		EndMS:       in.EndMS,
		Status:      "draft",
		ExportPath:  nil,
	})
}

func (s *Service) Get(ctx context.Context, userID int64, id int64) (clipsrepo.Clip, error) {
	c, err := s.clipsRepo.GetByIDForUser(ctx, userID, id)
	if err != nil {
		if errors.Is(err, clipsrepo.ErrNotFound) {
			return clipsrepo.Clip{}, ErrNotFound
		}
		return clipsrepo.Clip{}, err
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, userID int64, recordingUUID *string) ([]clipsrepo.Clip, error) {
	var recordingID *int64
	if recordingUUID != nil && *recordingUUID != "" {
		rec, err := s.recordingsRepo.GetByUUIDForUser(ctx, userID, *recordingUUID, 0)
		if err != nil {
			return nil, ErrNotFound
		}
		recordingID = &rec.ID
	}

	return s.clipsRepo.ListByUser(ctx, userID, recordingID)
}

type UpdateInput struct {
	Title   *string
	Caption *string
	StartMS *int
	EndMS   *int
	Status  *string
}

func (s *Service) Update(ctx context.Context, userID int64, id int64, in UpdateInput) (clipsrepo.Clip, error) {
	if in.StartMS != nil && in.EndMS != nil && *in.EndMS <= *in.StartMS {
		return clipsrepo.Clip{}, ErrBadInput
	}

	c, err := s.clipsRepo.UpdateByIDForUser(ctx, userID, id, clipsrepo.UpdateParams{
		Title:   in.Title,
		Caption: in.Caption,
		StartMS: in.StartMS,
		EndMS:   in.EndMS,
		Status:  in.Status,
	})
	if err != nil {
		if errors.Is(err, clipsrepo.ErrNotFound) {
			return clipsrepo.Clip{}, ErrNotFound
		}
		return clipsrepo.Clip{}, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, userID int64, id int64) error {
	err := s.clipsRepo.DeleteByIDForUser(ctx, userID, id)
	if err != nil {
		if errors.Is(err, clipsrepo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Export generates an mp4 file using ffmpeg and updates export_path + status.
// NOTE: Set FFMPEG_PATH to either:
//   - the directory containing ffmpeg.exe (Windows), OR
//   - the full path to ffmpeg.exe
func (s *Service) Export(ctx context.Context, userID int64, id int64) (clipsrepo.Clip, error) {
	// Fail early with a clearer error if ffmpeg isn't resolvable.
	if s.ffmpegPath == "ffmpeg" || s.ffmpegPath == "ffmpeg.exe" {
		// try one more time at runtime
		s.ffmpegPath = resolveFFmpegPath()
	}
	if strings.EqualFold(filepath.Base(s.ffmpegPath), "ffmpeg") || strings.EqualFold(filepath.Base(s.ffmpegPath), "ffmpeg.exe") {
		if _, err := exec.LookPath(s.ffmpegPath); err != nil && !filepath.IsAbs(s.ffmpegPath) {
			return clipsrepo.Clip{}, fmt.Errorf(`ffmpeg not found. Add it to PATH or set env FFMPEG_PATH (e.g. setx FFMPEG_PATH "D:\tools\ffmpeg...\bin" or full ffmpeg.exe path): %w`, err)
		}
	}

	c, err := s.clipsRepo.GetByIDForUser(ctx, userID, id)
	if err != nil {
		if errors.Is(err, clipsrepo.ErrNotFound) {
			return clipsrepo.Clip{}, ErrNotFound
		}
		return clipsrepo.Clip{}, err
	}

	inputPath, err := s.recordingsRepo.GetStoragePathByIDForUser(ctx, userID, c.RecordingID)
	if err != nil {
		return clipsrepo.Clip{}, ErrNotFound
	}
	if _, err := os.Stat(inputPath); err != nil {
		return clipsrepo.Clip{}, fmt.Errorf("recording file not found at %q: %w", inputPath, err)
	}

	if err := os.MkdirAll(s.clipsDir, 0755); err != nil {
		return clipsrepo.Clip{}, err
	}

	outPath := filepath.Join(s.clipsDir, fmt.Sprintf("clip_%d.mp4", c.ID))

	startSec := float64(c.StartMS) / 1000.0
	durSec := float64(c.EndMS-c.StartMS) / 1000.0
	if durSec <= 0 {
		return clipsrepo.Clip{}, ErrBadInput
	}

	// Use -t (duration) instead of -to (end time) to avoid ambiguity.
	cmd := exec.CommandContext(ctx, s.ffmpegPath,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", fmt.Sprintf("%.3f", startSec),
		"-t", fmt.Sprintf("%.3f", durSec),
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		outPath,
	)

	// Make PATH explicit for the child process (helps some Windows shells/IDEs).
	cmd.Env = os.Environ()

	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		failed := "failed"
		_, _ = s.clipsRepo.UpdateByIDForUser(ctx, userID, id, clipsrepo.UpdateParams{
			Status: &failed,
		})

		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = runErr.Error()
		}
		return clipsrepo.Clip{}, fmt.Errorf("ffmpeg failed: %s", msg)
	}

	ready := "ready"
	updated, err := s.clipsRepo.UpdateByIDForUser(ctx, userID, id, clipsrepo.UpdateParams{
		Status:     &ready,
		ExportPath: &outPath,
	})
	if err != nil {
		return clipsrepo.Clip{}, err
	}

	return updated, nil
}
