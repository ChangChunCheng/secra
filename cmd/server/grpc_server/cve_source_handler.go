package grpc_server

import (
	"context"
	"time"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// CVESourceServiceHandler implements secra_v1.CVESourceServiceServer.
type CVESourceServiceHandler struct {
	secra_v1.UnimplementedCVESourceServiceServer
}

// CreateCVESource creates a new CVE source.
func (h *CVESourceServiceHandler) CreateCVESource(ctx context.Context, req *secra_v1.CreateCVESourceRequest) (*secra_v1.CVESource, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	srcRepo := repo.NewCVESourceRepo(db.DB)
	srcSvc := service.NewCveSourceService(srcRepo)

	in := req.GetSource()
	created, err := srcSvc.Create(ctx, in.GetName(), in.GetUrl(), in.GetType(), in.GetDescription())
	if err != nil {
		return nil, err
	}

	return &secra_v1.CVESource{
		Id:          created.ID,
		Name:        created.Name,
		Type:        created.Type,
		Url:         created.URL,
		Enabled:     created.Enabled,
		CreatedAt:   created.CreatedAt.Format(time.RFC3339),
		Description: created.Description,
	}, nil
}

// GetCVESource retrieves a CVE source by ID.
func (h *CVESourceServiceHandler) GetCVESource(ctx context.Context, req *secra_v1.GetCVESourceRequest) (*secra_v1.CVESource, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	srcRepo := repo.NewCVESourceRepo(db.DB)
	srcSvc := service.NewCveSourceService(srcRepo)

	fetched, err := srcSvc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &secra_v1.CVESource{
		Id:          fetched.ID,
		Name:        fetched.Name,
		Type:        fetched.Type,
		Url:         fetched.URL,
		Enabled:     fetched.Enabled,
		CreatedAt:   fetched.CreatedAt.Format(time.RFC3339),
		Description: fetched.Description,
	}, nil
}

// ListCVESource returns a paginated list of CVE sources.
func (h *CVESourceServiceHandler) ListCVESource(ctx context.Context, req *secra_v1.ListCVESourceRequest) (*secra_v1.ListCVESourceResponse, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	srcRepo := repo.NewCVESourceRepo(db.DB)
	srcSvc := service.NewCveSourceService(srcRepo)

	items, err := srcSvc.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}

	var sources []*secra_v1.CVESource
	for _, s := range items {
		sources = append(sources, &secra_v1.CVESource{
			Id:          s.ID,
			Name:        s.Name,
			Type:        s.Type,
			Url:         s.URL,
			Enabled:     s.Enabled,
			CreatedAt:   s.CreatedAt.Format(time.RFC3339),
			Description: s.Description,
		})
	}
	return &secra_v1.ListCVESourceResponse{
		Sources: sources,
		Total:   int32(len(sources)),
	}, nil
}

// UpdateCVESource modifies an existing CVE source.
func (h *CVESourceServiceHandler) UpdateCVESource(ctx context.Context, req *secra_v1.UpdateCVESourceRequest) (*secra_v1.CVESource, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	srcRepo := repo.NewCVESourceRepo(db.DB)
	srcSvc := service.NewCveSourceService(srcRepo)

	in := req.GetSource()
	modelSrc := &model.CVESource{
		ID:          in.GetId(),
		Name:        in.GetName(),
		Type:        in.GetType(),
		URL:         in.GetUrl(),
		Description: in.GetDescription(),
		Enabled:     in.GetEnabled(),
	}
	updated, err := srcSvc.Update(ctx, modelSrc)
	if err != nil {
		return nil, err
	}

	return &secra_v1.CVESource{
		Id:          updated.ID,
		Name:        updated.Name,
		Type:        updated.Type,
		Url:         updated.URL,
		Enabled:     updated.Enabled,
		CreatedAt:   updated.CreatedAt.Format(time.RFC3339),
		Description: updated.Description,
	}, nil
}

// DeleteCVESource removes a CVE source by ID.
func (h *CVESourceServiceHandler) DeleteCVESource(ctx context.Context, req *secra_v1.DeleteCVESourceRequest) (*emptypb.Empty, error) {
	cfg := config.Load()
	db := storage.NewDB(cfg.PostgresDSN, false)

	srcRepo := repo.NewCVESourceRepo(db.DB)
	srcSvc := service.NewCveSourceService(srcRepo)

	if err := srcSvc.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
