package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/auth"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServicer defines the interface for user operations.
type UserServicer interface {
	CheckPassword(password string, confirmPassword string) (*string, error)
	Register(ctx context.Context, username, email, password string, confirmPassword string) (*model.User, error)
	OAuthLogin(ctx context.Context, provider, providerUserID, email string) (*model.User, error)
	Login(ctx context.Context, username, password string) (string, int64, error)
	GetProfile(ctx context.Context, token string) (*model.User, error)
	UpdateProfile(ctx context.Context, token, email, password string, confirmPassword string) (*model.User, error)
}

// ensure UserService implements UserServicer
var _ UserServicer = (*UserService)(nil)

// UserService encapsulates user registration and OAuth login logic.
type UserService struct {
	repo *repo.UserRepository
}

type UserJWTSecret struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
}

// NewUserService creates a new UserService.
func NewUserService(r *repo.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) CheckPassword(password string, confirmPassword string) (*string, error) {
	passwordHash := ""
	var err error
	// validate password fields
	if password != "" || confirmPassword != "" {
		if password != confirmPassword {
			return nil, status.Errorf(codes.InvalidArgument, "password and confirmPassword must match")
		}
		// password hash
		passwordHash, err = auth.HashPassword(password)
		if err != nil {
			return nil, err
		}
	}
	return &passwordHash, nil
}

// Register creates a new system user.
func (s *UserService) Register(ctx context.Context, username, email, password string, confirmPassword string) (*model.User, error) {

	passwordHash, err := s.CheckPassword(password, confirmPassword)
	if err != nil {
		return nil, err
	}

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
		PasswordHash:       *passwordHash,
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
	if auth.CheckPasswordHash(password, user.PasswordHash) != nil {
		return "", 0, errors.New("invalid credentials")
	}

	sub, err := json.Marshal(UserJWTSecret{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
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
	token, err := tokenObj.SignedString(config.Load().JWTConfig.Secret)
	if err != nil {
		return "", 0, err
	}
	return token, expireAt, nil
}

func (s *UserService) parseJWT(token string) (*UserJWTSecret, error) {
	// parse token
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return config.Load().JWTConfig.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return nil, errors.New("invalid token claims")
	}
	// subject contains marshaled UserJWTSecret
	var secret *UserJWTSecret
	if err := json.Unmarshal([]byte(claims.Subject), &secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// GetProfile parses the JWT token, retrieves the user ID, and returns the user profile.
func (s *UserService) GetProfile(ctx context.Context, token string) (*model.User, error) {
	secret, err := s.parseJWT(token)
	if err != nil {
		return nil, err
	}
	// retrieve user by ID
	return s.repo.FindByID(ctx, secret.ID.String())
}

// UpdateProfile parses the JWT token, updates the user's email (ignoring fullName), and returns the updated user.
// Note: fullName field is not stored in the user model and is ignored.
func (s *UserService) UpdateProfile(ctx context.Context, token, email, password string, confirmPassword string) (*model.User, error) {

	passwordHash, err := s.CheckPassword(password, confirmPassword)
	if err != nil {
		return nil, err
	}

	// parse token to get user ID
	secret, err := s.parseJWT(token)
	if err != nil {
		return nil, err
	}
	// update email via repository
	return s.repo.UpdateEmailAndPassword(ctx, secret.ID.String(), email, *passwordHash)
}
