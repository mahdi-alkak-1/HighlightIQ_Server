package clipcandidates

import "time"

type Candidate struct {
	ID           int64
	RecordingID  int64
	StartMS      int
	EndMS        int
	Score        float64
	DetectedJSON *string // store raw JSON string (nullable)
	Status       string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateParams struct {
	RecordingID  int64
	StartMS      int
	EndMS        int
	Score        float64
	DetectedJSON *string
	Status       string
}
