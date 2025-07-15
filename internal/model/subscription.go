package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Subscription represents a user’s subscription settings.
type Subscription struct {
	bun.BaseModel `bun:"table:subscriptions,alias:subscription"`

	ID                uuid.UUID            `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID            uuid.UUID            `bun:"user_id,type:uuid,nullzero"`
	SeverityThreshold string               `bun:"severity_threshold,default:'LOW'"`
	CreatedAt         time.Time            `bun:"created_at,notnull,default:current_timestamp"`
	Targets           []SubscriptionTarget `bun:"rel:has-many,join:id=subscription_id"`
}

// SubscriptionTarget represents a target item under a subscription.
type SubscriptionTarget struct {
	bun.BaseModel `bun:"table:subscription_targets,alias:subscription_target"`

	ID             uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	SubscriptionID uuid.UUID `bun:"subscription_id,type:uuid"`
	TargetTypeID   int       `bun:"target_type_id"`
	TargetID       uuid.UUID `bun:"target_id,type:uuid"`
}
