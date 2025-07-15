package grpc_server

import (
	"context"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// VendorServiceHandler implements secra_v1.VendorServiceServer.
type VendorServiceHandler struct {
	secra_v1.UnimplementedVendorServiceServer
}

// CreateVendor creates a new vendor record.
func (h *VendorServiceHandler) CreateVendor(ctx context.Context, req *secra_v1.CreateVendorRequest) (*secra_v1.Vendor, error) {
	// load config and initialize DB
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	// create service
	vendorRepo := repo.NewVendorRepo(db.DB)
	vendorSvc := service.NewVendorService(vendorRepo)

	// create vendor
	src := req.GetVendor()
	v, err := vendorSvc.Create(ctx, src.GetName())
	if err != nil {
		return nil, err
	}

	// map model.Vendor to gRPC Vendor
	return &secra_v1.Vendor{
		Id:   v.ID,
		Name: v.Name,
	}, nil
}

// GetVendor fetches a vendor by ID.
func (h *VendorServiceHandler) GetVendor(ctx context.Context, req *secra_v1.GetVendorRequest) (*secra_v1.Vendor, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	vendorRepo := repo.NewVendorRepo(db.DB)
	vendorSvc := service.NewVendorService(vendorRepo)

	v, err := vendorSvc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &secra_v1.Vendor{Id: v.ID, Name: v.Name}, nil
}

// ListVendor returns paginated list of vendors.
func (h *VendorServiceHandler) ListVendor(ctx context.Context, req *secra_v1.ListVendorRequest) (*secra_v1.ListVendorResponse, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	vendorRepo := repo.NewVendorRepo(db.DB)
	vendorSvc := service.NewVendorService(vendorRepo)

	items, err := vendorSvc.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}

	var vendors []*secra_v1.Vendor
	for _, v := range items {
		vendors = append(vendors, &secra_v1.Vendor{Id: v.ID, Name: v.Name})
	}
	return &secra_v1.ListVendorResponse{Vendors: vendors, Total: int32(len(vendors))}, nil
}

// UpdateVendor is a stub. TODO: implement.
func (h *VendorServiceHandler) UpdateVendor(ctx context.Context, req *secra_v1.UpdateVendorRequest) (*secra_v1.Vendor, error) {
	// load config and initialize DB
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	// service
	vendorRepo := repo.NewVendorRepo(db.DB)
	vendorSvc := service.NewVendorService(vendorRepo)

	// update vendor
	src := req.GetVendor()
	modelV := &model.Vendor{ID: src.GetId(), Name: src.GetName()}
	updated, err := vendorSvc.Update(ctx, modelV)
	if err != nil {
		return nil, err
	}

	return &secra_v1.Vendor{Id: updated.ID, Name: updated.Name}, nil
}

// DeleteVendor is a stub. TODO: implement.
func (h *VendorServiceHandler) DeleteVendor(ctx context.Context, req *secra_v1.DeleteVendorRequest) (*emptypb.Empty, error) {
	// load config and initialize DB
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	// service
	vendorRepo := repo.NewVendorRepo(db.DB)
	vendorSvc := service.NewVendorService(vendorRepo)

	// delete vendor
	if err := vendorSvc.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
