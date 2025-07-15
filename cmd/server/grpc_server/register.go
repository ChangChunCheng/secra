package grpc_server

import (
	"google.golang.org/grpc"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
)

// RegisterServices registers all gRPC service handlers with the given server.
func RegisterServices(server *grpc.Server) {
	// Register CVE service
	secra_v1.RegisterCVEServiceServer(server, &CVEServiceHandler{})

	// Register Vendor service
	secra_v1.RegisterVendorServiceServer(server, &VendorServiceHandler{})

	// Register User service
	secra_v1.RegisterUserServiceServer(server, &UserServiceHandler{})
}
