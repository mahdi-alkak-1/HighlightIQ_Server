package auth

import (
	"context"

	authsvc "highlightiq-server/internal/services/auth"
)

type AuthService interface {
	Register(ctx context.Context, in authsvc.RegisterInput) (authsvc.RegisterOutput, error)
	Login(ctx context.Context, in authsvc.LoginInput) (authsvc.RegisterOutput, error)
}

type Handler struct {
	svc AuthService
}

func New(svc AuthService) *Handler {
	return &Handler{svc: svc}
}
