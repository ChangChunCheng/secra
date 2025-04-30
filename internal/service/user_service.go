// --- internal/service/user_service.go ---

package service

import (
	"context"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
)

type UserServiceServer struct {
	secra_v1.UnimplementedUserServiceServer
}

func NewUserServiceServer() *UserServiceServer {
	return &UserServiceServer{}
}

func (s *UserServiceServer) LocalLogin(ctx context.Context, req *secra_v1.LoginRequest) (*secra_v1.LoginResponse, error) {
	// TODO: implement local login authentication logic
	return &secra_v1.LoginResponse{
		Token:    "dummy-token",
		ExpireAt: "2099-12-31T23:59:59Z",
	}, nil
}

func (s *UserServiceServer) OAuthLogin(ctx context.Context, req *secra_v1.OAuthLoginRequest) (*secra_v1.LoginResponse, error) {
	// TODO: implement OAuth2 login authentication logic
	return &secra_v1.LoginResponse{
		Token:    "dummy-oauth-token",
		ExpireAt: "2099-12-31T23:59:59Z",
	}, nil
}
