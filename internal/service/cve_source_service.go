package service

import (
	"context"

	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// CveSourceService handles CVE resource operations.
type CveSourceService struct {
	repo *repo.CVESourceRepo
}

// NewCveSourceService creates a new CveSourceService.
func NewCveSourceService(r *repo.CVESourceRepo) *CveSourceService {
	return &CveSourceService{repo: r}
}

// Create adds a new CVE resource.
func (s *CveSourceService) Create(ctx context.Context, name, url, ctype, desc string) (*model.CVESource, error) {
	src := &model.CVESource{
		// ID is nil, DB will generate
		Type:        ctype,
		Name:        name,
		URL:         url,
		Description: desc,
	}
	if err := s.repo.Create(ctx, src); err != nil {
		return nil, err
	}
	return src, nil
}

// Get retrieves a CVE source by ID.
func (s *CveSourceService) Get(ctx context.Context, id string) (*model.CVESource, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a paginated list of CVE sources.
func (s *CveSourceService) List(ctx context.Context, limit, offset int) ([]model.CVESource, error) {
	return s.repo.List(ctx, limit, offset)
}

// Update modifies an existing CVE source.
func (s *CveSourceService) Update(ctx context.Context, src *model.CVESource) (*model.CVESource, error) {
	if err := s.repo.Update(ctx, src); err != nil {
		return nil, err
	}
	return src, nil
}

// Delete removes a CVE source by ID.
func (s *CveSourceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
