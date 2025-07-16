package repo

import (
	"context"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/model"
)

// SubscriptionRepository handles database operations for subscriptions.
type SubscriptionRepository struct {
	db *bun.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository.
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

// GetTargetTypes retrieves all target_types and returns a map[id]name.
func (r *SubscriptionRepository) GetTargetTypes(ctx context.Context) (map[int]string, error) {
	rows, err := r.db.NewSelect().
		Model((*model.TargetType)(nil)).
		ColumnExpr("id, name").
		TableExpr("target_types").
		Rows(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int]string)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[id] = name
	}
	return m, nil
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

// DeleteSubscription deletes a subscription and its targets for a user.
func (r *SubscriptionRepository) DeleteSubscription(ctx context.Context, userID string, subscriptionID string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// delete targets first
		if _, err := tx.NewDelete().
			Model((*model.SubscriptionTarget)(nil)).
			Where("subscription_id = ?", subscriptionID).
			Exec(ctx); err != nil {
			return err
		}
		// delete subscription, ensure it belongs to user
		if _, err := tx.NewDelete().
			Model((*model.Subscription)(nil)).
			Where("id = ?", subscriptionID).
			Where("user_id = ?", userID).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	})
}
