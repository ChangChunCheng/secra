package grpc_server

import (
	"context"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// VendorServiceHandler implements secra_v1.VendorServiceServer.
type VendorServiceHandler struct {
	secra_v1.UnimplementedVendorServiceServer
}

// CreateVendor is a stub. TODO: implement.
func (h *VendorServiceHandler) CreateVendor(ctx context.Context, req *secra_v1.CreateVendorRequest) (*secra_v1.Vendor, error) {
	// TODO: implement CreateVendor
	return nil, nil
}

// GetVendor is a stub. TODO: implement.
func (h *VendorServiceHandler) GetVendor(ctx context.Context, req *secra_v1.GetVendorRequest) (*secra_v1.Vendor, error) {
	// TODO: implement GetVendor
	return nil, nil
}

// ListVendor is a stub. TODO: implement.
func (h *VendorServiceHandler) ListVendor(ctx context.Context, req *secra_v1.ListVendorRequest) (*secra_v1.ListVendorResponse, error) {
	// TODO: implement ListVendor
	return nil, nil
}

// UpdateVendor is a stub. TODO: implement.
func (h *VendorServiceHandler) UpdateVendor(ctx context.Context, req *secra_v1.UpdateVendorRequest) (*secra_v1.Vendor, error) {
	// TODO: implement UpdateVendor
	return nil, nil
}

// DeleteVendor is a stub. TODO: implement.
func (h *VendorServiceHandler) DeleteVendor(ctx context.Context, req *secra_v1.DeleteVendorRequest) (*emptypb.Empty, error) {
	// TODO: implement DeleteVendor
	return nil, nil
}
