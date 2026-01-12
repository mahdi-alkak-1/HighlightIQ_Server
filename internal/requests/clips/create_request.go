package clips

type CreateRequest struct {
	RecordingUUID string `json:"recording_uuid" validate:"required"`
	CandidateID   *int64 `json:"candidate_id" validate:"omitempty,gt=0"`

	Title   string  `json:"title" validate:"required,min=1,max=120"`
	Caption *string `json:"caption" validate:"omitempty,max=5000"`

	StartMS int `json:"start_ms" validate:"gte=0"`
	EndMS   int `json:"end_ms" validate:"gt=0"`
}

func (r CreateRequest) Validate() error {
	return validate.Struct(r)
}
