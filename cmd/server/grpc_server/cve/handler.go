package cve

import (
	"context"
	"time"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/service"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements secra_v1.CVEServiceServer.
type Handler struct {
	secra_v1.UnimplementedCVEServiceServer
	cveService service.CveServicer
}

// NewHandler creates a new Handler.
func NewHandler(svc service.CveServicer) *Handler {
	return &Handler{cveService: svc}
}

// CreateCVE creates a new CVE entry.
func (h *Handler) CreateCVE(ctx context.Context, req *secra_v1.CreateCVERequest) (*secra_v1.CVE, error) {
	in := req.GetCve()
	created, err := h.cveService.Create(ctx,
		"", // let service generate ID
		in.GetSourceId(),
		in.GetSourceUid(),
		in.GetTitle(),
		in.GetDescription(),
	)
	if err != nil {
		return nil, err
	}

	out := &secra_v1.CVE{
		Id:          created.ID,
		SourceId:    created.SourceID,
		SourceUid:   created.SourceUID,
		Title:       created.Title,
		Description: created.Description,
		Severity:    "",
		CvssScore:   0,
		Status:      created.Status,
		PublishedAt: created.PublishedAt.Format(time.RFC3339),
		UpdatedAt:   created.UpdatedAt.Format(time.RFC3339),
	}
	if created.Severity != nil {
		out.Severity = *created.Severity
	}
	if created.CVSSScore != nil {
		out.CvssScore = float32(*created.CVSSScore)
	}
	return out, nil
}

// GetCVE retrieves a CVE by its ID.
func (h *Handler) GetCVE(ctx context.Context, req *secra_v1.GetCVERequest) (*secra_v1.CVE, error) {
	fetched, err := h.cveService.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	out := &secra_v1.CVE{
		Id:          fetched.ID,
		SourceId:    fetched.SourceID,
		SourceUid:   fetched.SourceUID,
		Title:       fetched.Title,
		Description: fetched.Description,
		Severity:    "",
		CvssScore:   0,
		Status:      fetched.Status,
		PublishedAt: fetched.PublishedAt.Format(time.RFC3339),
		UpdatedAt:   fetched.UpdatedAt.Format(time.RFC3339),
	}
	if fetched.Severity != nil {
		out.Severity = *fetched.Severity
	}
	if fetched.CVSSScore != nil {
		out.CvssScore = float32(*fetched.CVSSScore)
	}
	return out, nil
}

// ListCVE returns a list of CVEs with pagination.
func (h *Handler) ListCVE(ctx context.Context, req *secra_v1.ListCVERequest) (*secra_v1.ListCVEResponse, error) {
	items, err := h.cveService.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}

	var cves []*secra_v1.CVE
	for _, c := range items {
		entry := &secra_v1.CVE{
			Id:          c.ID,
			SourceId:    c.SourceID,
			SourceUid:   c.SourceUID,
			Title:       c.Title,
			Description: c.Description,
			Severity:    "",
			CvssScore:   0,
			Status:      c.Status,
			PublishedAt: c.PublishedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}
		if c.Severity != nil {
			entry.Severity = *c.Severity
		}
		if c.CVSSScore != nil {
			entry.CvssScore = float32(*c.CVSSScore)
		}
		cves = append(cves, entry)
	}

	return &secra_v1.ListCVEResponse{
		Cves:  cves,
		Total: int32(len(cves)),
	}, nil
}

// UpdateCVE modifies an existing CVE.
func (h *Handler) UpdateCVE(ctx context.Context, req *secra_v1.UpdateCVERequest) (*secra_v1.CVE, error) {
	in := req.GetCve()
	updated, err := h.cveService.Update(ctx, in)
	if err != nil {
		return nil, err
	}

	out := &secra_v1.CVE{
		Id:          updated.ID,
		SourceId:    updated.SourceID,
		SourceUid:   updated.SourceUID,
		Title:       updated.Title,
		Description: updated.Description,
		Severity:    "",
		CvssScore:   0,
		Status:      updated.Status,
		PublishedAt: updated.PublishedAt.Format(time.RFC3339),
		UpdatedAt:   updated.UpdatedAt.Format(time.RFC3339),
	}
	if updated.Severity != nil {
		out.Severity = *updated.Severity
	}
	if updated.CVSSScore != nil {
		out.CvssScore = float32(*updated.CVSSScore)
	}
	return out, nil
}

// DeleteCVE removes a CVE by ID.
func (h *Handler) DeleteCVE(ctx context.Context, req *secra_v1.DeleteCVERequest) (*emptypb.Empty, error) {
	if err := h.cveService.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
