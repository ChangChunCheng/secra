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
func (s *CveSourceService) Create(ctx context.Context, name, url string) (*model.CVESource, error) {
	src := &model.CVESource{
		// ID is nil, DB will generate
		Name: name,
		URL:  url,
	}
	if err := s.repo.Create(ctx, src); err != nil {
		return nil, err
	}
	return src, nil
}
