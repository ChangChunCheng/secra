package grpc_server

import (
	"context"
	"time"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
)

// UserServiceHandler implements secra_v1.UserServiceServer.
type UserServiceHandler struct {
	secra_v1.UnimplementedUserServiceServer
}

// Register is a stub. TODO: implement.
func (h *UserServiceHandler) Register(ctx context.Context, req *secra_v1.RegisterRequest) (*secra_v1.RegisterResponse, error) {
	// load config and initialize DB
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	// create service
	userRepo := repo.NewUserRepository(db.DB)
	userSvc := service.NewUserService(userRepo)

	// register user
	_, err := userSvc.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	// 僅回應註冊成功
	return &secra_v1.RegisterResponse{
		Message: "Create Success",
	}, nil
}

// Login is a stub. TODO: implement.
func (h *UserServiceHandler) Login(ctx context.Context, req *secra_v1.LoginRequest) (*secra_v1.LoginResponse, error) {
	// load config and initialize DB
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	// create service
	userRepo := repo.NewUserRepository(db.DB)
	userSvc := service.NewUserService(userRepo)

	// authenticate user
	token, expireAt, err := userSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	t := time.Unix(expireAt, 0)

	// 回傳 JWT token
	return &secra_v1.LoginResponse{
		Token:    token,
		ExpireAt: t.Format("2006-01-02T15:04:05"),
	}, nil
}

// GetProfile is a stub. TODO: implement.
func (h *UserServiceHandler) GetProfile(ctx context.Context, req *secra_v1.TokenRequest) (*secra_v1.UserProfile, error) {
	// load config and initialize DB
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	// create service
	userRepo := repo.NewUserRepository(db.DB)
	userSvc := service.NewUserService(userRepo)

	// get user profile from token
	usr, err := userSvc.GetProfile(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	// map model.User to UserProfile
	return &secra_v1.UserProfile{
		Id:       usr.ID.String(),
		Username: usr.Username,
		Email:    usr.Email,
	}, nil
}

// UpdateProfile is a stub. TODO: implement.
func (h *UserServiceHandler) UpdateProfile(ctx context.Context, req *secra_v1.UpdateProfileRequest) (*secra_v1.UserProfile, error) {
	// load config and initialize DB
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	// create service
	userRepo := repo.NewUserRepository(db.DB)
	userSvc := service.NewUserService(userRepo)

	// update user profile (only email; fullName ignored)
	usr, err := userSvc.UpdateProfile(ctx, req.Token, req.Email)
	if err != nil {
		return nil, err
	}

	// map model.User to UserProfile
	return &secra_v1.UserProfile{
		Id:       usr.ID.String(),
		Username: usr.Username,
		Email:    usr.Email,
	}, nil
}
