package youtubepublishes

import "time"

type CreateInput struct {
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

type UpdateInput struct {
	YoutubeURL   *string
	Status       *string
	PublishedAt  *time.Time
	LastSyncedAt *time.Time
	Views        *int
	Likes        *int
	Comments     *int
	Analytics    *string
}
