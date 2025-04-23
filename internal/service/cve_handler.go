package service

import (
	"context"
	"time"

	"gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *SecraHandler) CreateCVESource(ctx context.Context, req *secra_v1.CreateCVESourceRequest) (*secra_v1.CVESource, error) {
	r := repo.NewCVESourceRepo(s.DB)
	m := &model.CVESource{
		ID:      req.Source.Id,
		Name:    req.Source.Name,
		Type:    req.Source.Type,
		URL:     &req.Source.Url,
		Enabled: req.Source.Enabled,
	}
	if err := r.Create(ctx, m); err != nil {
		return nil, err
	}
	return req.Source, nil
}

func (s *SecraHandler) GetCVESource(ctx context.Context, req *secra_v1.GetCVESourceRequest) (*secra_v1.CVESource, error) {
	r := repo.NewCVESourceRepo(s.DB)
	m, err := r.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &secra_v1.CVESource{
		Id:        m.ID,
		Name:      m.Name,
		Type:      m.Type,
		Url:       derefString(m.URL),
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *SecraHandler) UpdateCVESource(ctx context.Context, req *secra_v1.UpdateCVESourceRequest) (*secra_v1.CVESource, error) {
	r := repo.NewCVESourceRepo(s.DB)
	m := &model.CVESource{
		ID:      req.Source.Id,
		Name:    req.Source.Name,
		Type:    req.Source.Type,
		URL:     &req.Source.Url,
		Enabled: req.Source.Enabled,
	}
	if err := r.Update(ctx, m); err != nil {
		return nil, err
	}
	return req.Source, nil
}

func (s *SecraHandler) ListCVESource(ctx context.Context, req *secra_v1.ListCVESourceRequest) (*secra_v1.ListCVESourceResponse, error) {
	r := repo.NewCVESourceRepo(s.DB)
	items, err := r.List(ctx, int(req.Page.Limit), int(req.Page.Offset))
	if err != nil {
		return nil, err
	}
	var out []*secra_v1.CVESource
	for _, v := range items {
		out = append(out, &secra_v1.CVESource{
			Id:        v.ID,
			Name:      v.Name,
			Type:      v.Type,
			Url:       derefString(v.URL),
			Enabled:   v.Enabled,
			CreatedAt: v.CreatedAt.Format(time.RFC3339),
		})
	}
	return &secra_v1.ListCVESourceResponse{
		Sources: out,
		Page:    &secra_v1.PageResponse{Total: int32(len(out))},
	}, nil
}

func (s *SecraHandler) DeleteCVESource(ctx context.Context, req *secra_v1.DeleteCVESourceRequest) (*emptypb.Empty, error) {
	r := repo.NewCVESourceRepo(s.DB)
	if err := r.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
