package vendor

import (
	"context"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/service"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements secra_v1.VendorServiceServer.
type Handler struct {
	secra_v1.UnimplementedVendorServiceServer
	vendorService service.VendorServicer
}

// NewHandler creates a new Handler.
func NewHandler(svc service.VendorServicer) *Handler {
	return &Handler{vendorService: svc}
}

// CreateVendor creates a new vendor record.
func (h *Handler) CreateVendor(ctx context.Context, req *secra_v1.CreateVendorRequest) (*secra_v1.Vendor, error) {
	src := req.GetVendor()
	v, err := h.vendorService.Create(ctx, src.GetName())
	if err != nil {
		return nil, err
	}

	return &secra_v1.Vendor{
		Id:   v.ID,
		Name: v.Name,
	}, nil
}

// GetVendor fetches a vendor by ID.
func (h *Handler) GetVendor(ctx context.Context, req *secra_v1.GetVendorRequest) (*secra_v1.Vendor, error) {
	v, err := h.vendorService.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &secra_v1.Vendor{Id: v.ID, Name: v.Name}, nil
}

// ListVendor returns paginated list of vendors.
func (h *Handler) ListVendor(ctx context.Context, req *secra_v1.ListVendorRequest) (*secra_v1.ListVendorResponse, error) {
	items, err := h.vendorService.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
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
func (h *Handler) UpdateVendor(ctx context.Context, req *secra_v1.UpdateVendorRequest) (*secra_v1.Vendor, error) {
	src := req.GetVendor()
	modelV := &model.Vendor{ID: src.GetId(), Name: src.GetName()}
	updated, err := h.vendorService.Update(ctx, modelV)
	if err != nil {
		return nil, err
	}

	return &secra_v1.Vendor{Id: updated.ID, Name: updated.Name}, nil
}

// DeleteVendor is a stub. TODO: implement.
func (h *Handler) DeleteVendor(ctx context.Context, req *secra_v1.DeleteVendorRequest) (*emptypb.Empty, error) {
	if err := h.vendorService.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
