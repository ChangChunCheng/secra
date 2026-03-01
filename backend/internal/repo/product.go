package repo

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

type ProductRepo struct {
	db *bun.DB
}

func NewProductRepo(db *bun.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	_, err := r.db.NewInsert().Model(p).Exec(ctx)
	return err
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*model.Product, error) {
	p := new(model.Product)
	err := r.db.NewSelect().Model(p).Where("id = ?", id).Scan(ctx)
	return p, err
}

func (r *ProductRepo) List(ctx context.Context, limit, offset int) ([]model.Product, error) {
	var list []model.Product
	err := r.db.NewSelect().Model(&list).
		Limit(limit).
		Offset(offset).
		Order("name ASC").
		Scan(ctx)
	return list, err
}

func (r *ProductRepo) Update(ctx context.Context, p *model.Product) error {
	_, err := r.db.NewUpdate().Model(p).WherePK().Exec(ctx)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Product)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
