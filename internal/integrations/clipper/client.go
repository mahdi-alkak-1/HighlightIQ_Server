package clipper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 2 * time.Hour,
		},
	}
}

// ---- KILL DETECTION TYPES ----

type DetectKillsRequest struct {
	Path string `json:"path"`

	// clip rules
	MaxClipSeconds    int     `json:"max_clip_seconds,omitempty"`
	PreRollSeconds    int     `json:"pre_roll_seconds,omitempty"`
	PostRollSeconds   int     `json:"post_roll_seconds,omitempty"`
	MinClipSeconds    int     `json:"min_clip_seconds,omitempty"`
	MinSpacingSeconds float64 `json:"min_spacing_seconds,omitempty"`

	// scan speed/quality
	SampleFPS float64 `json:"sample_fps,omitempty"`

	// merge nearby events into same clip
	MergeGapSeconds int `json:"merge_gap_seconds,omitempty"`

	// detection tuning (optional)
	ElimMatchThreshold float64 `json:"elim_match_threshold,omitempty"`
	MinConsecutiveHits int     `json:"min_consecutive_hits,omitempty"`
	CooldownSeconds    float64 `json:"cooldown_seconds,omitempty"`
}

type Candidate struct {
	StartMS int     `json:"start_ms"`
	EndMS   int     `json:"end_ms"`
	Score   float64 `json:"score"`
}

type DetectKillsResponse struct {
	Candidates      []Candidate `json:"candidates"`
	VideoEndSeconds float64     `json:"video_end_seconds"`
	KillsDetected   int         `json:"kills_detected"`
}

func (c *Client) DetectKills(ctx context.Context, req DetectKillsRequest) (DetectKillsResponse, error) {
	var out DetectKillsResponse

	body, err := json.Marshal(req)
	if err != nil {
		return out, fmt.Errorf("marshal detect-kills request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/detect-kills", bytes.NewReader(body))
	if err != nil {
		return out, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	res, err := c.http.Do(httpReq)
	if err != nil {
		return out, fmt.Errorf("call clipper: %w", err)
	}
	defer res.Body.Close()

	// Better error detail if Python returns an error payload
	if res.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10)) // 8KB
		if len(b) > 0 {
			return out, fmt.Errorf("clipper returned status %d: %s", res.StatusCode, string(b))
		}
		return out, fmt.Errorf("clipper returned status %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return out, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}
