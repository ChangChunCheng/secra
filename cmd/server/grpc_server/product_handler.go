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

// ProductServiceHandler implements secra_v1.ProductServiceServer.
type ProductServiceHandler struct {
	secra_v1.UnimplementedProductServiceServer
}

// CreateProduct creates a new product record.
func (h *ProductServiceHandler) CreateProduct(ctx context.Context, req *secra_v1.CreateProductRequest) (*secra_v1.Product, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	productRepo := repo.NewProductRepo(db.DB)
	productSvc := service.NewProductService(productRepo)

	src := req.GetProduct()
	p, err := productSvc.Create(ctx, src.GetVendorId(), src.GetName())
	if err != nil {
		return nil, err
	}

	return &secra_v1.Product{
		Id:       p.ID,
		Name:     p.Name,
		VendorId: p.VendorID,
	}, nil
}

// GetProduct fetches a product by ID.
func (h *ProductServiceHandler) GetProduct(ctx context.Context, req *secra_v1.GetProductRequest) (*secra_v1.Product, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	productRepo := repo.NewProductRepo(db.DB)
	productSvc := service.NewProductService(productRepo)

	p, err := productSvc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &secra_v1.Product{
		Id:       p.ID,
		Name:     p.Name,
		VendorId: p.VendorID,
	}, nil
}

// ListProduct returns paginated list of products.
func (h *ProductServiceHandler) ListProduct(ctx context.Context, req *secra_v1.ListProductRequest) (*secra_v1.ListProductResponse, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	productRepo := repo.NewProductRepo(db.DB)
	productSvc := service.NewProductService(productRepo)

	items, err := productSvc.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}

	var products []*secra_v1.Product
	for _, p := range items {
		products = append(products, &secra_v1.Product{
			Id:       p.ID,
			Name:     p.Name,
			VendorId: p.VendorID,
		})
	}

	return &secra_v1.ListProductResponse{
		Products: products,
		Total:    int32(len(products)),
	}, nil
}

// UpdateProduct updates an existing product.
func (h *ProductServiceHandler) UpdateProduct(ctx context.Context, req *secra_v1.UpdateProductRequest) (*secra_v1.Product, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	productRepo := repo.NewProductRepo(db.DB)
	productSvc := service.NewProductService(productRepo)

	src := req.GetProduct()
	p := &model.Product{
		ID:       src.GetId(),
		VendorID: src.GetVendorId(),
		Name:     src.GetName(),
	}
	updated, err := productSvc.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return &secra_v1.Product{
		Id:       updated.ID,
		Name:     updated.Name,
		VendorId: updated.VendorID,
	}, nil
}

// DeleteProduct deletes a product by its ID.
func (h *ProductServiceHandler) DeleteProduct(ctx context.Context, req *secra_v1.DeleteProductRequest) (*emptypb.Empty, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	productRepo := repo.NewProductRepo(db.DB)
	productSvc := service.NewProductService(productRepo)

	if err := productSvc.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
