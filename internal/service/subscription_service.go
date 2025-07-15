package service

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// SubscriptionService handles business logic for subscriptions.
type SubscriptionService struct {
	repo *repo.SubscriptionRepository
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(r *repo.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: r}
}

// CreateSubscription creates a subscription with its targets.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, userID string, targets []model.SubscriptionTarget, severity string) (*model.Subscription, error) {
	sub := &model.Subscription{
		UserID:            mustParseUUID(userID),
		SeverityThreshold: severity,
	}
	if err := s.repo.CreateSubscription(ctx, sub, targets); err != nil {
		return nil, err
	}
	return sub, nil
}

// ListSubscriptions returns all subscriptions for a user.
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID string) ([]model.Subscription, error) {
	return s.repo.GetSubscriptionsByUser(ctx, userID)
}

// helper to parse string to uuid.UUID
func mustParseUUID(s string) uuid.UUID {
	u, _ := uuid.Parse(s)
	return u
}
