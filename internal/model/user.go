// internal/model/user.go

package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Username           string    `bun:"username,unique,notnull" json:"username"`
	Email              string    `bun:"email,unique,notnull" json:"email"`
	PasswordHash       string    `bun:"password_hash,notnull" json:"-"`
	Role               string    `bun:"role,notnull" json:"role"`     // "user" or "admin"
	Status             string    `bun:"status,notnull" json:"status"` // "active" or "disabled"
	MustChangePassword bool      `bun:"must_change_password,notnull,default:false" json:"must_change_password"`
	OAuthProvider      *string   `bun:"oauth_provider,nullzero" json:"oauth_provider,omitempty"`
	OAuthID            *string   `bun:"oauth_id,nullzero" json:"oauth_id,omitempty"`
	CreatedAt          time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
