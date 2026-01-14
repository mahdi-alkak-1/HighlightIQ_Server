package youtubepublishes

import "time"

type YoutubePublish struct {
	ID             int64      `json:"id"`
	ClipID         int64      `json:"clip_id"`
	YoutubeVideoID string     `json:"youtube_video_id"`
	YoutubeURL     string     `json:"youtube_url"`
	Status         string     `json:"status"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	LastSyncedAt   *time.Time `json:"last_synced_at,omitempty"`
	Views          int        `json:"views"`
	Likes          int        `json:"likes"`
	Comments       int        `json:"comments"`
	Analytics      *string    `json:"analytics,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateParams struct {
	ClipID         int64
	YoutubeVideoID string
	YoutubeURL     string
	Status         string
	PublishedAt    *time.Time
	LastSyncedAt   *time.Time
	Views          int
	Likes          int
	Comments       int
	Analytics      *string
}

type UpdateParams struct {
	YoutubeURL   *string
	Status       *string
	PublishedAt  *time.Time
	LastSyncedAt *time.Time
	Views        *int
	Likes        *int
	Comments     *int
	Analytics    *string
}
