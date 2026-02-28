package cvesource

import (
	"context"
	"time"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/service"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements secra_v1.CVESourceServiceServer.
type Handler struct {
	secra_v1.UnimplementedCVESourceServiceServer
	cveSourceService service.CveSourceServicer
}

// NewHandler creates a new Handler.
func NewHandler(svc service.CveSourceServicer) *Handler {
	return &Handler{cveSourceService: svc}
}

// CreateCVESource creates a new CVE source.
func (h *Handler) CreateCVESource(ctx context.Context, req *secra_v1.CreateCVESourceRequest) (*secra_v1.CVESource, error) {
	in := req.GetSource()
	created, err := h.cveSourceService.Create(ctx, in.GetName(), in.GetUrl(), in.GetType(), in.GetDescription())
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
func (h *Handler) GetCVESource(ctx context.Context, req *secra_v1.GetCVESourceRequest) (*secra_v1.CVESource, error) {
	fetched, err := h.cveSourceService.Get(ctx, req.GetId())
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
func (h *Handler) ListCVESource(ctx context.Context, req *secra_v1.ListCVESourceRequest) (*secra_v1.ListCVESourceResponse, error) {
	items, err := h.cveSourceService.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
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
func (h *Handler) UpdateCVESource(ctx context.Context, req *secra_v1.UpdateCVESourceRequest) (*secra_v1.CVESource, error) {
	in := req.GetSource()
	modelSrc := &model.CVESource{
		ID:          in.GetId(),
		Name:        in.GetName(),
		Type:        in.GetType(),
		URL:         in.GetUrl(),
		Description: in.GetDescription(),
		Enabled:     in.GetEnabled(),
	}
	updated, err := h.cveSourceService.Update(ctx, modelSrc)
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
func (h *Handler) DeleteCVESource(ctx context.Context, req *secra_v1.DeleteCVESourceRequest) (*emptypb.Empty, error) {
	if err := h.cveSourceService.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
