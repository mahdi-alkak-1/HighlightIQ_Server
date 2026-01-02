package clipcandidates

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=new approved rejected"`
}

func (r UpdateStatusRequest) Validate() error {
	return validate.Struct(r)
}
