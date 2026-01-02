package clipper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
			Timeout: 60 * time.Second, // video scan can take time
		},
	}
}

type DetectRequest struct {
	Path              string  `json:"path"`
	ClipLengthSeconds int     `json:"clip_length_seconds"`
	Threshold         float64 `json:"threshold"`
	MinClipSeconds    int     `json:"min_clip_seconds"`
}

type Candidate struct {
	StartMS int     `json:"start_ms"`
	EndMS   int     `json:"end_ms"`
	Score   float64 `json:"score"`
}

type DetectResponse struct {
	Candidates      []Candidate `json:"candidates"`
	VideoEndSeconds float64     `json:"video_end_seconds"`
	ScenesDetected  int         `json:"scenes_detected"`
}

func (c *Client) DetectCandidates(ctx context.Context, req DetectRequest) (DetectResponse, error) {
	var out DetectResponse

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/detect-candidates", bytes.NewReader(body))
	if err != nil {
		return out, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	res, err := c.http.Do(httpReq)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return out, fmt.Errorf("clipper returned status %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}
