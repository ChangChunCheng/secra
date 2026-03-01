package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx)
	return user, err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	return user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	return user, err
}

func (r *UserRepository) UpdateFullProfile(ctx context.Context, u *model.User) error {
	_, err := r.db.NewUpdate().Model(u).
		Column("email", "password_hash", "notification_frequency", "timezone", "updated_at").
		WherePK().
		Exec(ctx)
	return err
}

func (r *UserRepository) UpdateEmailAndPassword(ctx context.Context, userID, email, passwordHash string) (*model.User, error) {
	user := &model.User{ID: uuid.MustParse(userID), Email: email, PasswordHash: passwordHash}
	_, err := r.db.NewUpdate().
		Model(user).
		Column("email", "password_hash", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil { return nil, err }
	return r.FindByID(ctx, userID)
}
