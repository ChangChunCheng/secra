package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

// UserRepository handles user persistence.
type UserRepository struct {
	db *bun.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user record.
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

// GetUserByUsername retrieves a user by username.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().
		Model(user).
		Where("username = ?", username).
		Scan(ctx)
	return user, err
}

// GetUserByEmail retrieves a user by email.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().
		Model(user).
		Where("email = ?", email).
		Scan(ctx)
	return user, err
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	return user, err
}

// UpdateEmail updates the user's email.
func (r *UserRepository) UpdateEmail(ctx context.Context, id, email string) (*model.User, error) {
	user := &model.User{ID: uuid.MustParse(id), Email: email}
	_, err := r.db.NewUpdate().
		Model(user).
		Column("email", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}
