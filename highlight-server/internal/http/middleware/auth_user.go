package middleware

import "context"

type ctxKey string

const authUserKey ctxKey = "auth_user"

type AuthUser struct {
	ID    int64
	UUID  string
	Email string
}

func WithAuthUser(ctx context.Context, u AuthUser) context.Context {
	return context.WithValue(ctx, authUserKey, u)
}

func GetAuthUser(ctx context.Context) (AuthUser, bool) {
	v := ctx.Value(authUserKey)
	u, ok := v.(AuthUser)
	return u, ok
}
