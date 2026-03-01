package repo

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

type CVERepo struct {
	db *bun.DB
}

func NewCVERepo(db *bun.DB) *CVERepo {
	return &CVERepo{db: db}
}

func (r *CVERepo) Create(ctx context.Context, c *model.CVE) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	return err
}

func (r *CVERepo) GetByID(ctx context.Context, id string) (*model.CVE, error) {
	c := new(model.CVE)
	err := r.db.NewSelect().Model(c).Where("id = ?", id).Scan(ctx)
	return c, err
}

func (r *CVERepo) List(ctx context.Context, limit, offset int) ([]model.CVE, error) {
	var list []model.CVE
	err := r.db.NewSelect().Model(&list).
		Limit(limit).
		Offset(offset).
		Order("published_at DESC").
		Scan(ctx)
	return list, err
}

func (r *CVERepo) Update(ctx context.Context, c *model.CVE) error {
	_, err := r.db.NewUpdate().Model(c).WherePK().Exec(ctx)
	return err
}

func (r *CVERepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.CVE)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
