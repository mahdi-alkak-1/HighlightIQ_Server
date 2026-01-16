package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"highlightiq-server/internal/http/response"
	"highlightiq-server/internal/repos/users"
)

type JWTAuth struct {
	usersRepo *users.Repo
	secret    []byte
}

func NewJWTAuth(usersRepo *users.Repo, jwtSecret string) *JWTAuth {
	return &JWTAuth{
		usersRepo: usersRepo,
		secret:    []byte(jwtSecret),
	}
}

func (a *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "missing authorization header"})
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(h, prefix) {
			response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "invalid authorization header"})
			return
		}

		raw := strings.TrimSpace(strings.TrimPrefix(h, prefix))
		if raw == "" {
			response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "invalid authorization header"})
			return
		}

		token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
			// Ensure HS256
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return a.secret, nil
		})
		if err != nil || token == nil || !token.Valid {
			response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "invalid token claims"})
			return
		}

		sub, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)
		if sub == "" {
			response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "invalid token claims"})
			return
		}

		u, err := a.usersRepo.GetByUUID(r.Context(), sub)
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				response.JSON(w, http.StatusUnauthorized, map[string]any{"message": "invalid token user"})
				return
			}
			response.JSON(w, http.StatusInternalServerError, map[string]any{"message": "internal server error"})
			return
		}

		ctx := WithAuthUser(r.Context(), AuthUser{
			ID:    u.ID,
			UUID:  u.UUID,
			Email: email,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
