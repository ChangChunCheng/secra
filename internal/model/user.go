package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID                 uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Username           string    `bun:"username,notnull,unique" json:"username"`
	Email              string    `bun:"email,notnull,unique" json:"email"`
	PasswordHash       string    `bun:"password_hash,notnull" json:"-"`
	Role               string    `bun:"role,notnull,default:'user'" json:"role"`
	Status             string    `bun:"status,notnull,default:'active'" json:"status"`
	MustChangePassword bool      `bun:"must_change_password,default:false" json:"must_change_password"`
	
	// Preferences
	NotificationFrequency string    `bun:"notification_frequency,default:'daily'" json:"notification_frequency"`
	Timezone              string    `bun:"timezone,default:'UTC'" json:"timezone"`
	LastNotifiedAt        time.Time `bun:"last_notified_at" json:"last_notified_at"`

	OAuthProvider      *string   `bun:"oauth_provider" json:"oauth_provider,omitempty"`
	OAuthID            *string   `bun:"oauth_id" json:"oauth_id,omitempty"`
	CreatedAt          time.Time `bun:"created_at,nullzero,notnull,default:now()" json:"created_at"`
	UpdatedAt          time.Time `bun:"updated_at,nullzero,notnull,default:now()" json:"updated_at"`
}
