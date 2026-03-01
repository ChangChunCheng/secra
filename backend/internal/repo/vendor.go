package repo

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

type VendorRepo struct {
	db *bun.DB
}

func NewVendorRepo(db *bun.DB) *VendorRepo {
	return &VendorRepo{db: db}
}

func (r *VendorRepo) Create(ctx context.Context, v *model.Vendor) error {
	_, err := r.db.NewInsert().Model(v).Exec(ctx)
	return err
}

func (r *VendorRepo) GetByID(ctx context.Context, id string) (*model.Vendor, error) {
	v := new(model.Vendor)
	err := r.db.NewSelect().Model(v).Where("id = ?", id).Scan(ctx)
	return v, err
}

func (r *VendorRepo) List(ctx context.Context, limit, offset int) ([]model.Vendor, error) {
	var list []model.Vendor
	err := r.db.NewSelect().Model(&list).
		Limit(limit).
		Offset(offset).
		Order("name ASC").
		Scan(ctx)
	return list, err
}

func (r *VendorRepo) Update(ctx context.Context, v *model.Vendor) error {
	_, err := r.db.NewUpdate().Model(v).WherePK().Exec(ctx)
	return err
}

func (r *VendorRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Vendor)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
