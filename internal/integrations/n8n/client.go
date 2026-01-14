package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	clipsrepo "highlightiq-server/internal/repos/clips"
)

type Client struct {
	webhookURL string
	authHeader string
	http       *http.Client
}

func New(webhookURL string, authHeader string) *Client {
	return &Client{
		webhookURL: webhookURL,
		authHeader: authHeader,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type PublishPayload struct {
	ClipID      int64    `json:"clip_id"`
	ClipURL     string   `json:"clip_url"`
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	Ratio       string   `json:"ratio,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (c *Client) NotifyClipExported(ctx context.Context, clip clipsrepo.Clip, clipURL string) error {
	if c == nil || c.webhookURL == "" {
		return nil
	}

	payload := PublishPayload{
		ClipID:      clip.ID,
		ClipURL:     clipURL,
		Title:       clip.Title,
		Description: clip.Caption,
		Ratio:       "",
		Tags:        nil,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal n8n payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create n8n request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call n8n webhook: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
		if len(b) > 0 {
			return fmt.Errorf("n8n returned status %d: %s", res.StatusCode, string(b))
		}
		return fmt.Errorf("n8n returned status %d", res.StatusCode)
	}

	return nil
}
