package repo

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

type CVESourceRepo struct {
	db *bun.DB
}

func NewCVESourceRepo(db *bun.DB) *CVESourceRepo {
	return &CVESourceRepo{db: db}
}

func (r *CVESourceRepo) Create(ctx context.Context, src *model.CVESource) error {
	_, err := r.db.NewInsert().Model(src).Exec(ctx)
	return err
}

func (r *CVESourceRepo) GetByID(ctx context.Context, id string) (*model.CVESource, error) {
	src := new(model.CVESource)
	err := r.db.NewSelect().Model(src).Where("id = ?", id).Scan(ctx)
	return src, err
}

func (r *CVESourceRepo) List(ctx context.Context, limit, offset int) ([]model.CVESource, error) {
	var list []model.CVESource
	err := r.db.NewSelect().Model(&list).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Scan(ctx)
	return list, err
}

func (r *CVESourceRepo) Update(ctx context.Context, src *model.CVESource) error {
	_, err := r.db.NewUpdate().Model(src).
		WherePK().Exec(ctx)
	return err
}

func (r *CVESourceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.CVESource)(nil)).
		Where("id = ?", id).Exec(ctx)
	return err
}
