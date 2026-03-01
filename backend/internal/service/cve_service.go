package service

import (
	"context"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// CveServicer defines the interface for CVE operations.
type CveServicer interface {
	Create(ctx context.Context, _, sourceID, sourceUID, title, description string) (*model.CVE, error)
	Get(ctx context.Context, id string) (*model.CVE, error)
	List(ctx context.Context, limit, offset int) ([]*model.CVE, error)
	Update(ctx context.Context, in *secra_v1.CVE) (*model.CVE, error)
	Delete(ctx context.Context, id string) error
}

// ensure CveService implements CveServicer
var _ CveServicer = (*CveService)(nil)


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

// Get retrieves a CVE by ID.
func (s *CveService) Get(ctx context.Context, id string) (*model.CVE, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a list of CVEs with pagination.
func (s *CveService) List(ctx context.Context, limit, offset int) ([]*model.CVE, error) {
	items, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]*model.CVE, len(items))
	for i := range items {
		result[i] = &items[i]
	}
	return result, nil
}

// Update modifies an existing CVE record.
func (s *CveService) Update(ctx context.Context, in *secra_v1.CVE) (*model.CVE, error) {
	// map proto CVE -> model.CVE
	c := &model.CVE{
		ID:          in.GetId(),
		SourceID:    in.GetSourceId(),
		SourceUID:   in.GetSourceUid(),
		Title:       in.GetTitle(),
		Description: in.GetDescription(),
		Status:      in.GetStatus(),
	}
	if sev := in.GetSeverity(); sev != "" {
		c.Severity = &sev
	}
	if score := in.GetCvssScore(); score != 0 {
		f := float64(score)
		c.CVSSScore = &f
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Delete removes a CVE by ID.
func (s *CveService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
