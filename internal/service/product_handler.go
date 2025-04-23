package service

import (
	"context"

	"gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *SecraHandler) CreateProduct(ctx context.Context, req *secra_v1.CreateProductRequest) (*secra_v1.Product, error) {
	r := repo.NewProductRepo(s.DB)
	m := &model.Product{ID: req.Product.Id, Name: req.Product.Name, VendorID: req.Product.VendorId}
	if err := r.Create(ctx, m); err != nil {
		return nil, err
	}
	return req.Product, nil
}

func (s *SecraHandler) GetProduct(ctx context.Context, req *secra_v1.GetProductRequest) (*secra_v1.Product, error) {
	r := repo.NewProductRepo(s.DB)
	m, err := r.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &secra_v1.Product{Id: m.ID, Name: m.Name, VendorId: m.VendorID}, nil
}

func (s *SecraHandler) UpdateProduct(ctx context.Context, req *secra_v1.UpdateProductRequest) (*secra_v1.Product, error) {
	r := repo.NewProductRepo(s.DB)
	m := &model.Product{ID: req.Product.Id, Name: req.Product.Name, VendorID: req.Product.VendorId}
	if err := r.Update(ctx, m); err != nil {
		return nil, err
	}
	return req.Product, nil
}

func (s *SecraHandler) ListProduct(ctx context.Context, req *secra_v1.ListProductRequest) (*secra_v1.ListProductResponse, error) {
	r := repo.NewProductRepo(s.DB)
	items, err := r.List(ctx, int(req.Page.Limit), int(req.Page.Offset))
	if err != nil {
		return nil, err
	}
	var out []*secra_v1.Product
	for _, p := range items {
		out = append(out, &secra_v1.Product{Id: p.ID, Name: p.Name, VendorId: p.VendorID})
	}
	return &secra_v1.ListProductResponse{
		Products: out,
		Page:     &secra_v1.PageResponse{Total: int32(len(out))},
	}, nil
}

func (s *SecraHandler) DeleteProduct(ctx context.Context, req *secra_v1.DeleteProductRequest) (*emptypb.Empty, error) {
	r := repo.NewProductRepo(s.DB)
	if err := r.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
