package youtubepublishes

import "time"

type InternalMarkDeletedRequest struct {
	YoutubeVideoID string     `json:"youtube_video_id" validate:"required,max=32"`
	LastSyncedAt   *time.Time `json:"last_synced_at" validate:"omitempty"`
}

func (r InternalMarkDeletedRequest) Validate() error {
	return validate.Struct(r)
}
