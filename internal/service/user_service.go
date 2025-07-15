package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// UserService encapsulates user registration and OAuth login logic.
type UserService struct {
	repo *repo.UserRepository
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

func (s *UserService) Login(ctx context.Context, username, password string) (string, error) {
	// 驗證使用者
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user.PasswordHash != password {
		return "", errors.New("invalid credentials")
	}
	// 簽發 JWT（使用簡易範例，實際應使用安全金鑰和過期時間）
	token := uuid.New().String()
	return token, nil
}
