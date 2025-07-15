package grpc_server

import (
	"context"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// CVEServiceHandler implements secra_v1.CVEServiceServer.
type CVEServiceHandler struct {
	secra_v1.UnimplementedCVEServiceServer
}

// CreateCVE is a stub. TODO: implement.
func (h *CVEServiceHandler) CreateCVE(ctx context.Context, req *secra_v1.CreateCVERequest) (*secra_v1.CVE, error) {
	// TODO: implement CreateCVE
	return nil, nil
}

// GetCVE is a stub. TODO: implement.
func (h *CVEServiceHandler) GetCVE(ctx context.Context, req *secra_v1.GetCVERequest) (*secra_v1.CVE, error) {
	// TODO: implement GetCVE
	return nil, nil
}

// ListCVE is a stub. TODO: implement.
func (h *CVEServiceHandler) ListCVE(ctx context.Context, req *secra_v1.ListCVERequest) (*secra_v1.ListCVEResponse, error) {
	// TODO: implement ListCVE
	return nil, nil
}

// UpdateCVE is a stub. TODO: implement.
func (h *CVEServiceHandler) UpdateCVE(ctx context.Context, req *secra_v1.UpdateCVERequest) (*secra_v1.CVE, error) {
	// TODO: implement UpdateCVE
	return nil, nil
}

// DeleteCVE is a stub. TODO: implement.
func (h *CVEServiceHandler) DeleteCVE(ctx context.Context, req *secra_v1.DeleteCVERequest) (*emptypb.Empty, error) {
	// TODO: implement DeleteCVE
	return nil, nil
}
