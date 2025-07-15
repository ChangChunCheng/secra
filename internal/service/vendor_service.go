package service

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// VendorService encapsulates vendor creation and management logic.
type VendorService struct {
	repo *repo.VendorRepo
}

// NewVendorService creates a new VendorService.
func NewVendorService(r *repo.VendorRepo) *VendorService {
	return &VendorService{repo: r}
}

// Create creates a new vendor with the given name.
func (s *VendorService) Create(ctx context.Context, name string) (*model.Vendor, error) {
	v := &model.Vendor{
		ID:   uuid.New().String(),
		Name: name,
	}
	if err := s.repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VendorService) Get(ctx context.Context, id string) (*model.Vendor, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *VendorService) List(ctx context.Context, limit, offset int) ([]model.Vendor, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *VendorService) Update(ctx context.Context, v *model.Vendor) (*model.Vendor, error) {
	if err := s.repo.Update(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VendorService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
