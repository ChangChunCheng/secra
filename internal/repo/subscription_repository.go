package repo

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

type SubscriptionRepository struct {
	db *bun.DB
}

func NewSubscriptionRepository(db *bun.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// CreateSubscription inserts a subscription and its targets in a transaction.
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *model.Subscription, targets []model.SubscriptionTarget) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(sub).Exec(ctx); err != nil {
			return err
		}
		for _, t := range targets {
			t.SubscriptionID = sub.ID
			if _, err := tx.NewInsert().Model(&t).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetSubscriptionsByUser returns subscriptions and their targets for a user.
func (r *SubscriptionRepository) GetSubscriptionsByUser(ctx context.Context, userID string) ([]model.Subscription, error) {
	var subs []model.Subscription
	err := r.db.NewSelect().
		Model(&subs).
		Relation("Targets").
		Where("subscription.user_id = ?", userID).
		Scan(ctx)
	return subs, err
}
