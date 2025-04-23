package service

import (
	"context"

	"gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *SecraHandler) CreateVendor(ctx context.Context, req *secra_v1.CreateVendorRequest) (*secra_v1.Vendor, error) {
	r := repo.NewVendorRepo(s.DB)
	m := &model.Vendor{ID: req.Vendor.Id, Name: req.Vendor.Name}
	if err := r.Create(ctx, m); err != nil {
		return nil, err
	}
	return req.Vendor, nil
}

func (s *SecraHandler) GetVendor(ctx context.Context, req *secra_v1.GetVendorRequest) (*secra_v1.Vendor, error) {
	r := repo.NewVendorRepo(s.DB)
	m, err := r.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &secra_v1.Vendor{Id: m.ID, Name: m.Name}, nil
}

func (s *SecraHandler) UpdateVendor(ctx context.Context, req *secra_v1.UpdateVendorRequest) (*secra_v1.Vendor, error) {
	r := repo.NewVendorRepo(s.DB)
	m := &model.Vendor{ID: req.Vendor.Id, Name: req.Vendor.Name}
	if err := r.Update(ctx, m); err != nil {
		return nil, err
	}
	return req.Vendor, nil
}

func (s *SecraHandler) ListVendor(ctx context.Context, req *secra_v1.ListVendorRequest) (*secra_v1.ListVendorResponse, error) {
	r := repo.NewVendorRepo(s.DB)
	items, err := r.List(ctx, int(req.Page.Limit), int(req.Page.Offset))
	if err != nil {
		return nil, err
	}
	var out []*secra_v1.Vendor
	for _, v := range items {
		out = append(out, &secra_v1.Vendor{Id: v.ID, Name: v.Name})
	}
	return &secra_v1.ListVendorResponse{
		Vendors: out,
		Page:    &secra_v1.PageResponse{Total: int32(len(out))},
	}, nil
}

func (s *SecraHandler) DeleteVendor(ctx context.Context, req *secra_v1.DeleteVendorRequest) (*emptypb.Empty, error) {
	r := repo.NewVendorRepo(s.DB)
	if err := r.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
