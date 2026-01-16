package clips

import "time"

type Clip struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	RecordingID     int64     `json:"recording_id"`
	CandidateID     *int64    `json:"candidate_id,omitempty"`
	Title           string    `json:"title"`
	Caption         *string   `json:"caption,omitempty"`
	StartMS         int       `json:"start_ms"`
	EndMS           int       `json:"end_ms"`
	DurationSeconds int       `json:"duration_seconds"`
	Status          string    `json:"status"`
	ExportPath      *string   `json:"export_path,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateParams struct {
	UserID      int64
	RecordingID int64
	CandidateID *int64
	Title       string
	Caption     *string
	StartMS     int
	EndMS       int
	Status      string
	ExportPath  *string
}

type UpdateParams struct {
	Title      *string
	Caption    *string
	StartMS    *int
	EndMS      *int
	Status     *string
	ExportPath *string
}
