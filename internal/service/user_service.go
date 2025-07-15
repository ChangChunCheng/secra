package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// UserService encapsulates user registration and OAuth login logic.
type UserService struct {
	repo *repo.UserRepository
}

type UserJWTSecret struct {
	id       uuid.UUID
	username string
	role     string
}

// NewUserService creates a new UserService.
func NewUserService(r *repo.UserRepository) *UserService {
	return &UserService{repo: r}
}

// Register creates a new system user.
func (s *UserService) Register(ctx context.Context, username, email, passwordHash string) (*model.User, error) {
	// 防止重複
	if existing, _ := s.repo.GetUserByUsername(ctx, username); existing.ID != uuid.Nil {
		return nil, errors.New("username already exists")
	}
	if existing, _ := s.repo.GetUserByEmail(ctx, email); existing.ID != uuid.Nil {
		return nil, errors.New("email already exists")
	}
	user := &model.User{
		ID:                 uuid.New(),
		Username:           username,
		Email:              email,
		PasswordHash:       passwordHash,
		Role:               "user",
		Status:             "active",
		MustChangePassword: false,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// OAuthLogin finds or creates a user via OAuth.
func (s *UserService) OAuthLogin(ctx context.Context, provider, providerUserID, email string) (*model.User, error) {
	// existing code...
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil && user.ID != uuid.Nil {
		return user, nil
	}
	// otherwise create new
	newUser := &model.User{
		ID:                 uuid.New(),
		Username:           email,
		Email:              email,
		PasswordHash:       "",
		Role:               "user",
		Status:             "active",
		MustChangePassword: false,
		OAuthProvider:      &provider,
		OAuthID:            &providerUserID,
	}
	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (string, int64, error) {
	// 驗證使用者
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", 0, err
	}
	if user.PasswordHash != password {
		return "", 0, errors.New("invalid credentials")
	}

	sub, err := json.Marshal(UserJWTSecret{
		id:       user.ID,
		username: user.Username,
		role:     user.Role,
	})
	if err != nil {
		return "", 0, err
	}
	token, expireAt, err := s.generateJWT(sub)
	return token, expireAt, err
}

// generateJWT 建立 JWT 並回傳 token 及過期 Unix 時間
func (s *UserService) generateJWT(sub []byte) (string, int64, error) {

	expireAt := time.Now().Add(config.Load().JWTConfig.Expiry).Unix()
	claims := jwt.RegisteredClaims{
		Subject:   string(sub),
		ExpiresAt: jwt.NewNumericDate(time.Unix(expireAt, 0)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenObj.SignedString([]byte(config.Load().JWTConfig.Secret))
	if err != nil {
		return "", 0, err
	}
	return token, expireAt, nil
}
