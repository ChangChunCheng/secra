package user

import (
	"context"
	"fmt"

	userv1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/service"
)

type Handler struct {
	userv1.UnimplementedUserServiceServer
	userService service.UserServicer
}

func NewHandler(userService service.UserServicer) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	_, err := h.userService.Register(ctx, req.Username, req.Email, req.Password, req.ConfirmPassword)
	if err != nil {
		return nil, err
	}
	return &userv1.RegisterResponse{
		Message: "User registered successfully",
	}, nil
}

func (h *Handler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	token, expireAt, err := h.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &userv1.LoginResponse{
		Token:    token,
		ExpireAt: fmt.Sprintf("%d", expireAt),
	}, nil
}

func (h *Handler) GetProfile(ctx context.Context, req *userv1.TokenRequest) (*userv1.UserProfile, error) {
	user, err := h.userService.GetProfile(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return &userv1.UserProfile{
		Id:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (h *Handler) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UserProfile, error) {
	user, err := h.userService.UpdateProfile(ctx, req.Token, req.Email, req.Password, req.ConfirmPassword, "daily", "UTC")
	if err != nil {
		return nil, err
	}
	return &userv1.UserProfile{
		Id:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
