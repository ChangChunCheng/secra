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

func (r *SubscriptionRepository) GetTargetTypes(ctx context.Context) (map[int]string, error) {
	var results []model.TargetType
	err := r.db.NewSelect().Model(&results).Scan(ctx)
	m := make(map[int]string)
	for _, res := range results {
		m[res.ID] = res.Name
	}
	return m, err
}

func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *model.Subscription, targets []model.SubscriptionTarget) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(sub).Exec(ctx); err != nil {
			return err
		}
		for i := range targets {
			targets[i].SubscriptionID = sub.ID
		}
		if _, err := tx.NewInsert().Model(&targets).Exec(ctx); err != nil {
			return err
		}
		return nil
	})
}

func (r *SubscriptionRepository) GetSubscriptionsByUser(ctx context.Context, userID string) ([]model.Subscription, error) {
	var subs []model.Subscription
	err := r.db.NewSelect().Model(&subs).Where("user_id = ?", userID).Scan(ctx)
	return subs, err
}

func (r *SubscriptionRepository) DeleteSubscription(ctx context.Context, userID string, subscriptionID string) error {
	_, err := r.db.NewDelete().Table("subscriptions").
		Where("id = ? AND user_id = ?", subscriptionID, userID).
		Exec(ctx)
	return err
}

func (r *SubscriptionRepository) UpdateThreshold(ctx context.Context, userID string, subscriptionID string, threshold int16) error {
	_, err := r.db.NewUpdate().Table("subscriptions").
		Set("severity_threshold = ?", threshold).
		Where("id = ? AND user_id = ?", subscriptionID, userID).
		Exec(ctx)
	return err
}
