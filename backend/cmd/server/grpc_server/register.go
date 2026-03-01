package grpc_server

import (
	"github.com/uptrace/bun"
	"google.golang.org/grpc"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/cmd/server/grpc_server/cve"
	"gitlab.com/jacky850509/secra/cmd/server/grpc_server/cvesource"
	"gitlab.com/jacky850509/secra/cmd/server/grpc_server/product"
	"gitlab.com/jacky850509/secra/cmd/server/grpc_server/subscription"
	"gitlab.com/jacky850509/secra/cmd/server/grpc_server/user"
	"gitlab.com/jacky850509/secra/cmd/server/grpc_server/vendor"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
)

// RegisterServices registers all gRPC service handlers with the given server.
func RegisterServices(server *grpc.Server, db *bun.DB) {
	// --- CVE Service ---
	cveRepo := repo.NewCVERepo(db)
	cveSvc := service.NewCveService(cveRepo)
	cveHandler := cve.NewHandler(cveSvc)
	secra_v1.RegisterCVEServiceServer(server, cveHandler)

	// --- Product Service ---
	productRepo := repo.NewProductRepo(db)
	productSvc := service.NewProductService(productRepo)
	productHandler := product.NewHandler(productSvc)
	secra_v1.RegisterProductServiceServer(server, productHandler)

	// --- Vendor Service ---
	vendorRepo := repo.NewVendorRepo(db)
	vendorSvc := service.NewVendorService(vendorRepo)
	vendorHandler := vendor.NewHandler(vendorSvc)
	secra_v1.RegisterVendorServiceServer(server, vendorHandler)

	// --- CVE Source Service ---
	cveSourceRepo := repo.NewCVESourceRepo(db)
	cveSourceSvc := service.NewCveSourceService(cveSourceRepo)
	cveSourceHandler := cvesource.NewHandler(cveSourceSvc)
	secra_v1.RegisterCVESourceServiceServer(server, cveSourceHandler)

	// --- User Service ---
	userRepo := repo.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := user.NewHandler(userSvc)
	secra_v1.RegisterUserServiceServer(server, userHandler)

	// --- Subscription Service ---
	subscriptionRepo := repo.NewSubscriptionRepository(db)
	subscriptionSvc := service.NewSubscriptionService(subscriptionRepo)
	subscriptionHandler := subscription.NewHandler(subscriptionSvc)
	secra_v1.RegisterSubscriptionServiceServer(server, subscriptionHandler)
}
