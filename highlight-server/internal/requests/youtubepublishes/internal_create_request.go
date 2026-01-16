package youtubepublishes

import (
	"encoding/json"
	"time"
)

type InternalCreateRequest struct {
	ClipID         int64   `json:"clip_id" validate:"required,gt=0"`
	YoutubeVideoID string  `json:"youtube_video_id" validate:"required,max=32"`
	YoutubeURL     string  `json:"youtube_url" validate:"required,max=255"`
	Status         *string `json:"status" validate:"omitempty,oneof=queued uploaded failed"`

	PublishedAt  *time.Time `json:"published_at" validate:"omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at" validate:"omitempty"`

	Views    int `json:"views" validate:"gte=0"`
	Likes    int `json:"likes" validate:"gte=0"`
	Comments int `json:"comments" validate:"gte=0"`

	Analytics *json.RawMessage `json:"analytics,omitempty"`
}

func (r InternalCreateRequest) Validate() error {
	return validate.Struct(r)
}
