package auth

// ValidationError maps field -> error message
type ValidationError map[string]string

func (ValidationError) Error() string {
	return "validation failed"
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=60"`
	Email    string `json:"email" validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}
