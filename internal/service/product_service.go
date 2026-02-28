package service

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// ProductServicer defines the interface for product operations.
type ProductServicer interface {
	Create(ctx context.Context, vendorID, name string) (*model.Product, error)
	Get(ctx context.Context, id string) (*model.Product, error)
	List(ctx context.Context, limit, offset int) ([]model.Product, error)
	Update(ctx context.Context, p *model.Product) (*model.Product, error)
	Delete(ctx context.Context, id string) error
}

// ensure ProductService implements ProductServicer
var _ ProductServicer = (*ProductService)(nil)


// ProductService encapsulates product creation and management logic.
type ProductService struct {
	repo *repo.ProductRepo
}

// NewProductService creates a new ProductService.
func NewProductService(r *repo.ProductRepo) *ProductService {
	return &ProductService{repo: r}
}

// Create creates a new product with the given vendorID and name.
func (s *ProductService) Create(ctx context.Context, vendorID, name string) (*model.Product, error) {
	p := &model.Product{
		ID:       uuid.New().String(),
		VendorID: vendorID,
		Name:     name,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Get retrieves a product by its ID.
func (s *ProductService) Get(ctx context.Context, id string) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a list of products with pagination.
func (s *ProductService) List(ctx context.Context, limit, offset int) ([]model.Product, error) {
	return s.repo.List(ctx, limit, offset)
}

// Update updates an existing product.
func (s *ProductService) Update(ctx context.Context, p *model.Product) (*model.Product, error) {
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete deletes a product by its ID.
func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
