package clipcandidates

import (
	"strings"
)

type ValidationError map[string]string

func (v ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("validation error")
	return b.String()
}

type DetectRequest struct {
	ClipLengthSeconds int     `json:"clip_length_seconds" validate:"omitempty,min=5,max=120"`
	Threshold         float64 `json:"threshold" validate:"omitempty,gte=0,lte=100"`
	MinClipSeconds    int     `json:"min_clip_seconds" validate:"omitempty,min=1,max=60"`
	MaxCandidates     int     `json:"max_candidates" validate:"omitempty,min=1,max=200"`
	MinSpacingSeconds int     `json:"min_spacing_seconds" validate:"omitempty,min=0,max=60"`
}

func (r DetectRequest) Validate() error {
	if err := validate.Struct(r); err != nil {
		errs := ValidationError{}
		errs["general"] = "invalid request"
		return errs
	}
	return nil
}
