package repo

import (
	"context"

	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"

	"gitlab.com/jacky850509/secra/internal/model"
)

// UserRepo provides CRUD operations for users.
type UserRepo struct {
	db *bun.DB
}

// NewUserRepo returns a new UserRepo.
func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

// CreateLocalUser creates a user registered locally with username/email.
func (r *UserRepo) CreateLocalUser(ctx context.Context, username, email, password, role string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	u := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}
	_, err = r.db.NewInsert().Model(u).Exec(ctx)
	return err
}

// GetByUsername fetches a user by username.
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	u := new(model.User)
	err := r.db.NewSelect().Model(u).Where("username = ?", username).Scan(ctx)
	return u, err
}

// VerifyPassword compares the provided password with the stored hash.
func (r *UserRepo) VerifyPassword(u *model.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
