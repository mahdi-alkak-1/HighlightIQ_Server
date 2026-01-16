package recordings

type CreateInput struct {
	UserID       int64
	Title        string
	OriginalName string
	FileBytes    []byte
}

type RecordingDTO struct {
	UUID            string `json:"uuid"`
	Title           string `json:"title"`
	OriginalName    string `json:"original_filename"`
	StoragePath     string `json:"storage_path"`
	DurationSeconds int    `json:"duration_seconds"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
}

type UpdateTitleInput struct {
	UserID int64
	UUID   string
	Title  string
}
