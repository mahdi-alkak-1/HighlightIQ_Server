package clips

type UpdateRequest struct {
	Title   *string `json:"title" validate:"omitempty,min=1,max=120"`
	Caption *string `json:"caption" validate:"omitempty,max=5000"`

	StartMS *int `json:"start_ms" validate:"omitempty,gte=0"`
	EndMS   *int `json:"end_ms" validate:"omitempty,gt=0"`

	Status *string `json:"status" validate:"omitempty,oneof=draft ready published failed"`
}

func (r UpdateRequest) Validate() error {
	return validate.Struct(r)
}
