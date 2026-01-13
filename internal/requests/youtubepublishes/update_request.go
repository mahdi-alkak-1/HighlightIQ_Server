package youtubepublishes

import "time"

type UpdateRequest struct {
	YoutubeURL *string `json:"youtube_url" validate:"omitempty,max=255"`
	Status     *string `json:"status" validate:"omitempty,oneof=queued uploaded failed"`

	PublishedAt  *time.Time `json:"published_at" validate:"omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at" validate:"omitempty"`
}

func (r UpdateRequest) Validate() error {
	return validate.Struct(r)
}
