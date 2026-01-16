package youtubepublishes

import "time"

type CreateRequest struct {
	YoutubeVideoID string  `json:"youtube_video_id" validate:"required,max=32"`
	YoutubeURL     string  `json:"youtube_url" validate:"required,max=255"`
	Status         *string `json:"status" validate:"omitempty,oneof=queued uploaded failed"`

	PublishedAt  *time.Time `json:"published_at" validate:"omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at" validate:"omitempty"`
}

func (r CreateRequest) Validate() error {
	return validate.Struct(r)
}
