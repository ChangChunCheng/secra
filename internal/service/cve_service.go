package service

import (
	"context"

	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// CveService handles CVE operations.
type CveService struct {
	repo *repo.CVERepo
}

// NewCveService creates a new CveService.
func NewCveService(r *repo.CVERepo) *CveService {
	return &CveService{repo: r}
}

// Create adds a new CVE record.
func (s *CveService) Create(ctx context.Context, _, sourceID, sourceUID, title, description string) (*model.CVE, error) {
	cve := &model.CVE{
		SourceID:    sourceID,
		SourceUID:   sourceUID,
		Title:       title,
		Description: description,
	}
	if err := s.repo.Create(ctx, cve); err != nil {
		return nil, err
	}
	return cve, nil
}
