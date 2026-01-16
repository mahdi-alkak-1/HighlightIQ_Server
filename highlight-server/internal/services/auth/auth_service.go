package auth

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"highlightiq-server/internal/repos/users"
)

type Service struct {
	users     *users.Repo
	jwtSecret []byte
	tokenTTL  time.Duration
}

func New(usersRepo *users.Repo, jwtSecret string) *Service {
	return &Service{
		users:     usersRepo,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  24 * time.Hour,
	}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	// 1) Check if email already exists (nice error instead of raw SQL error)
	_, err := s.users.GetByEmail(ctx, in.Email)
	if err == nil {
		return RegisterOutput{}, ErrEmailTaken
	}
	if err != nil && !errors.Is(err, users.ErrNotFound) {
		return RegisterOutput{}, err
	}

	// 2) Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterOutput{}, err
	}

	// 3) Insert user
	u, err := s.users.Create(ctx, users.CreateParams{
		UUID:         uuid.NewString(),
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		// in case of race condition (two requests same email)
		if isDuplicateEmail(err) {
			return RegisterOutput{}, ErrEmailTaken
		}
		return RegisterOutput{}, err
	}

	// 4) Issue JWT
	token, err := s.signJWT(u.UUID, u.Email)
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{
		User: UserDTO{
			ID:    u.UUID,
			Name:  u.Name,
			Email: u.Email,
		},
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}

func isDuplicateEmail(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

func (s *Service) signJWT(userUUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userUUID,
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(s.tokenTTL).Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}

func (s *Service) Login(ctx context.Context, in LoginInput) (RegisterOutput, error) {
	// 1) Find user by email
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return RegisterOutput{}, ErrInvalidCredentials
		}
		return RegisterOutput{}, err
	}

	// 2) Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return RegisterOutput{}, ErrInvalidCredentials
	}

	// 3) Issue JWT (same signing as Register)
	token, err := s.signJWT(u.UUID, u.Email)
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{
		User: UserDTO{
			ID:    u.UUID,
			Name:  u.Name,
			Email: u.Email,
		},
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}
