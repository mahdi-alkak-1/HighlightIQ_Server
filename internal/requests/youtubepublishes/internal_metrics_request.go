package youtubepublishes

import (
	"encoding/json"
	"time"
)

type InternalMetricsRequest struct {
	YoutubeVideoID string `json:"youtube_video_id" validate:"required,max=32"`

	Views    *int `json:"views" validate:"omitempty,gte=0"`
	Likes    *int `json:"likes" validate:"omitempty,gte=0"`
	Comments *int `json:"comments" validate:"omitempty,gte=0"`

	PublishedAt  *time.Time       `json:"published_at" validate:"omitempty"`
	LastSyncedAt *time.Time       `json:"last_synced_at" validate:"omitempty"`
	Analytics    *json.RawMessage `json:"analytics,omitempty"`
}

func (r InternalMetricsRequest) Validate() error {
	return validate.Struct(r)
}
