package product

import (
	"context"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/service"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements secra_v1.ProductServiceServer.
type Handler struct {
	secra_v1.UnimplementedProductServiceServer
	productService service.ProductServicer
}

// NewHandler creates a new Handler.
func NewHandler(svc service.ProductServicer) *Handler {
	return &Handler{productService: svc}
}

// CreateProduct creates a new product record.
func (h *Handler) CreateProduct(ctx context.Context, req *secra_v1.CreateProductRequest) (*secra_v1.Product, error) {
	src := req.GetProduct()
	p, err := h.productService.Create(ctx, src.GetVendorId(), src.GetName())
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
func (h *Handler) GetProduct(ctx context.Context, req *secra_v1.GetProductRequest) (*secra_v1.Product, error) {
	p, err := h.productService.Get(ctx, req.GetId())
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
func (h *Handler) ListProduct(ctx context.Context, req *secra_v1.ListProductRequest) (*secra_v1.ListProductResponse, error) {
	items, err := h.productService.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
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
func (h *Handler) UpdateProduct(ctx context.Context, req *secra_v1.UpdateProductRequest) (*secra_v1.Product, error) {
	src := req.GetProduct()
	p := &model.Product{
		ID:       src.GetId(),
		VendorID: src.GetVendorId(),
		Name:     src.GetName(),
	}
	updated, err := h.productService.Update(ctx, p)
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
func (h *Handler) DeleteProduct(ctx context.Context, req *secra_v1.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := h.productService.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
