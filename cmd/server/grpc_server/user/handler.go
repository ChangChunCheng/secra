package user

import (
	"context"
	"time"

	"gitlab.com/jacky850509/secra/internal/service"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
)

// Handler implements secra_v1.UserServiceServer.
type Handler struct {
	secra_v1.UnimplementedUserServiceServer
	userService service.UserServicer
}

// NewHandler creates a new Handler.
func NewHandler(svc service.UserServicer) *Handler {
	return &Handler{userService: svc}
}

// Register is a stub. TODO: implement.
func (h *Handler) Register(ctx context.Context, req *secra_v1.RegisterRequest) (*secra_v1.RegisterResponse, error) {
	// register user
	_, err := h.userService.Register(ctx, req.Username, req.Email, req.Password, req.ConfirmPassword)
	if err != nil {
		return nil, err
	}

	// 僅回應註冊成功
	return &secra_v1.RegisterResponse{
		Message: "Create Success",
	},
		nil
}

// Login is a stub. TODO: implement.
func (h *Handler) Login(ctx context.Context, req *secra_v1.LoginRequest) (*secra_v1.LoginResponse, error) {
	// authenticate user
	token, expireAt, err := h.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	t := time.Unix(expireAt, 0)

	// 回傳 JWT token
	return &secra_v1.LoginResponse{
		Token:    token,
		ExpireAt: t.Format("2006-01-02T15:04:05"),
	},
		nil
}

// GetProfile is a stub. TODO: implement.
func (h *Handler) GetProfile(ctx context.Context, req *secra_v1.TokenRequest) (*secra_v1.UserProfile, error) {
	// get user profile from token
	usr, err := h.userService.GetProfile(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	// map model.User to UserProfile
	return &secra_v1.UserProfile{
		Id:       usr.ID.String(),
		Username: usr.Username,
		Email:    usr.Email,
	},
		nil
}

// UpdateProfile is a stub. TODO: implement.
func (h *Handler) UpdateProfile(ctx context.Context, req *secra_v1.UpdateProfileRequest) (*secra_v1.UserProfile, error) {
	// update user profile, include password if provided
	usr, err := h.userService.UpdateProfile(ctx, req.Token, req.Email, req.Password, req.ConfirmPassword)
	if err != nil {
		return nil, err
	}

	// map model.User to UserProfile
	return &secra_v1.UserProfile{
		Id:       usr.ID.String(),
		Username: usr.Username,
		Email:    usr.Email,
	},
		nil
}
