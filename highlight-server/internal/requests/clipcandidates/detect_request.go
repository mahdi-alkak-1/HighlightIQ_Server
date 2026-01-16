package clipcandidates

import "strings"

type ValidationError map[string]string

func (v ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("validation error")
	return b.String()
}

type DetectRequest struct {
	MaxClipSeconds    int     `json:"max_clip_seconds" validate:"omitempty,min=5,max=120"`
	PreRollSeconds    int     `json:"pre_roll_seconds" validate:"omitempty,min=0,max=30"`
	PostRollSeconds   int     `json:"post_roll_seconds" validate:"omitempty,min=0,max=30"`
	MinClipSeconds    int     `json:"min_clip_seconds" validate:"omitempty,min=1,max=60"`
	SampleFPS         float64 `json:"sample_fps" validate:"omitempty,gt=0,lte=60"`
	MinSpacingSeconds float64 `json:"min_spacing_seconds" validate:"omitempty,gte=0,lte=120"`

	MergeGapSeconds int `json:"merge_gap_seconds" validate:"omitempty,min=0,max=60"`

	ElimMatchThreshold float64 `json:"elim_match_threshold" validate:"omitempty,gte=-1,lte=1"`
	MinConsecutiveHits int     `json:"min_consecutive_hits" validate:"omitempty,min=1,max=60"`
	CooldownSeconds    float64 `json:"cooldown_seconds" validate:"omitempty,gte=0,lte=120"`
}

func (r DetectRequest) Validate() error {
	if err := validate.Struct(r); err != nil {
		errs := ValidationError{}
		errs["general"] = "invalid request"
		return errs
	}
	return nil
}
