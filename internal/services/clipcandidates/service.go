package clipcandidates

import (
	"context"
	"errors"
	"sort"

	"highlightiq-server/internal/integrations/clipper"
	candidatesrepo "highlightiq-server/internal/repos/clipcandidates"
	recordingsrepo "highlightiq-server/internal/repos/recordings"
)

var ErrNotFound = errors.New("clipcandidates: recording not found")

type Service struct {
	recordings *recordingsrepo.Repo
	candidates *candidatesrepo.Repo
	clipper    *clipper.Client
}

func New(recordings *recordingsrepo.Repo, candidates *candidatesrepo.Repo, clipperClient *clipper.Client) *Service {
	return &Service{
		recordings: recordings,
		candidates: candidates,
		clipper:    clipperClient,
	}
}

type DetectInput struct {
	RecordingUUID     string
	ClipLengthSeconds int
	Threshold         float64
	MinClipSeconds    int
	MaxCandidates     int
	MinSpacingSeconds int // avoid many near-duplicates
}

func (s *Service) DetectAndStore(ctx context.Context, userID int64, in DetectInput) (int64, error) {
	if in.RecordingUUID == "" {
		return 0, ErrNotFound
	}

	rec, err := s.recordings.GetByUUIDForUser(ctx, userID, in.RecordingUUID, 0)
	if err != nil {
		return 0, ErrNotFound
	}

	// defaults
	if in.ClipLengthSeconds <= 0 {
		in.ClipLengthSeconds = 30
	}
	if in.Threshold <= 0 {
		in.Threshold = 27
	}
	if in.MinClipSeconds <= 0 {
		in.MinClipSeconds = 10
	}
	if in.MaxCandidates <= 0 {
		in.MaxCandidates = 20
	}
	if in.MinSpacingSeconds <= 0 {
		in.MinSpacingSeconds = 8
	}

	resp, err := s.clipper.DetectCandidates(ctx, clipper.DetectRequest{
		Path:              rec.StoragePath,
		ClipLengthSeconds: in.ClipLengthSeconds,
		Threshold:         in.Threshold,
		MinClipSeconds:    in.MinClipSeconds,
	})
	if err != nil {
		return 0, err
	}

	// Sort by score desc, then start asc
	sort.Slice(resp.Candidates, func(i, j int) bool {
		if resp.Candidates[i].Score == resp.Candidates[j].Score {
			return resp.Candidates[i].StartMS < resp.Candidates[j].StartMS
		}
		return resp.Candidates[i].Score > resp.Candidates[j].Score
	})

	// spacing filter + cap
	minSpacingMS := in.MinSpacingSeconds * 1000
	var picked []clipper.Candidate
	for _, c := range resp.Candidates {
		ok := true
		for _, p := range picked {
			if abs(c.StartMS-p.StartMS) < minSpacingMS {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		picked = append(picked, c)
		if len(picked) >= in.MaxCandidates {
			break
		}
	}

	toInsert := make([]candidatesrepo.CreateParams, 0, len(picked))
	for _, c := range picked {
		toInsert = append(toInsert, candidatesrepo.CreateParams{
			RecordingID:  rec.ID,
			StartMS:      c.StartMS,
			EndMS:        c.EndMS,
			Score:        c.Score,
			DetectedJSON: nil,
			Status:       "new",
		})
	}

	return s.candidates.CreateMany(ctx, toInsert)
}

func (s *Service) ListByRecordingUUID(ctx context.Context, userID int64, recordingUUID string) ([]candidatesrepo.Candidate, error) {
	rec, err := s.recordings.GetByUUIDForUser(ctx, userID, recordingUUID, 0)
	if err != nil {
		return nil, ErrNotFound
	}
	return s.candidates.ListByRecordingID(ctx, rec.ID)
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, status string) error {
	return s.candidates.UpdateStatus(ctx, id, status)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.candidates.Delete(ctx, id)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
