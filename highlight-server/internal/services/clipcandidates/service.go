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
	RecordingUUID string

	// clip rules
	MaxClipSeconds  int
	PreRollSeconds  int
	PostRollSeconds int
	MinClipSeconds  int

	// scan speed/quality
	SampleFPS float64

	// picking rules
	MaxCandidates     int
	MinSpacingSeconds float64

	// merge kills close to each other into same clip
	MergeGapSeconds int

	ElimMatchThreshold float64
	MinConsecutiveHits int
	CooldownSeconds    float64

	// Optional tuning (if you want to expose them later)
	// CooldownSeconds is now exposed above
}

func (s *Service) DetectAndStore(ctx context.Context, userID int64, in DetectInput) (int64, error) {
	if in.RecordingUUID == "" {
		return 0, ErrNotFound
	}

	rec, err := s.recordings.GetByUUIDForUser(ctx, userID, in.RecordingUUID, 0)
	if err != nil {
		return 0, ErrNotFound
	}

	// ---- defaults (match python defaults) ----
	if in.MaxClipSeconds <= 0 {
		in.MaxClipSeconds = 60
	}
	if in.PreRollSeconds < 0 {
		in.PreRollSeconds = 0
	}
	if in.PreRollSeconds == 0 {
		in.PreRollSeconds = 5
	}
	if in.PostRollSeconds < 0 {
		in.PostRollSeconds = 0
	}
	if in.PostRollSeconds == 0 {
		in.PostRollSeconds = 3
	}
	if in.MinClipSeconds <= 0 {
		in.MinClipSeconds = 8
	}

	// scanning: banner detector likes ~60 fps
	if in.SampleFPS <= 0 {
		in.SampleFPS = 60.0
	}

	if in.MaxCandidates <= 0 {
		in.MaxCandidates = 20
	}

	// This is for spacing candidates after scoring; keep smaller so we don't throw away real multi-kill clips.
	if in.MinSpacingSeconds < 0 {
		in.MinSpacingSeconds = 0
	}
	if in.MinSpacingSeconds == 0 {
		in.MinSpacingSeconds = 2
	}

	if in.MergeGapSeconds < 0 {
		in.MergeGapSeconds = 0
	}
	if in.MergeGapSeconds == 0 {
		in.MergeGapSeconds = 0
	}

	if in.CooldownSeconds <= 0 {
		in.CooldownSeconds = 1.2
	}
	resp, err := s.clipper.DetectKills(ctx, clipper.DetectKillsRequest{
		Path:               rec.StoragePath,
		MaxClipSeconds:     in.MaxClipSeconds,
		PreRollSeconds:     in.PreRollSeconds,
		PostRollSeconds:    in.PostRollSeconds,
		MinClipSeconds:     in.MinClipSeconds,
		SampleFPS:          in.SampleFPS,
		MinSpacingSeconds:  in.MinSpacingSeconds,
		MergeGapSeconds:    in.MergeGapSeconds,
		ElimMatchThreshold: in.ElimMatchThreshold,
		MinConsecutiveHits: in.MinConsecutiveHits,
		CooldownSeconds:    in.CooldownSeconds,
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
	minSpacingMS := int(in.MinSpacingSeconds * 1000)
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
