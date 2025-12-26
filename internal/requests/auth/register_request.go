package authreq

import "github.com/go-playground/validator/v10"


var Validate = validator.New()

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=60"`
	Email    string `json:"email" validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}
