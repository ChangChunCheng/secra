package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
)

// SubscriptionServicer defines the interface for subscription operations.
type SubscriptionServicer interface {
	SeverityToString(level int16) string
	CreateSubscription(ctx context.Context, userID string, targets []SubscriptionTarget, severity string) (*model.Subscription, error)
	ListSubscriptions(ctx context.Context, userID string) ([]model.Subscription, error)
	DeleteSubscription(ctx context.Context, userID string, subscriptionID string) error
}

// ensure SubscriptionService implements SubscriptionServicer
var _ SubscriptionServicer = (*SubscriptionService)(nil)

// SubscriptionService handles business logic for subscriptions.
type SubscriptionService struct {
	repo          *repo.SubscriptionRepository
	targetTypeMap map[int]string
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(r *repo.SubscriptionRepository) *SubscriptionService {
	// preload target types map from repository
	m, err := r.GetTargetTypes(context.Background())
	if err != nil {
		// on error fallback to empty map
		m = make(map[int]string)
	}
	return &SubscriptionService{
		repo:          r,
		targetTypeMap: m,
	}
}

// SubscriptionTarget 用於 gRPC handler 與 service 間轉換
type SubscriptionTarget struct {
	TargetType string
	TargetID   string
}

// SeverityToString 將閾值轉回字串
func (s *SubscriptionService) SeverityToString(level int16) string {
	// lookup from preloaded map
	if name, ok := s.targetTypeMap[int(level)]; ok {
		return strings.ToUpper(name)
	}
	return "UNKNOWN"
}

// CreateSubscription creates a subscription with its targets.
func (s *SubscriptionService) CreateSubscription(ctx context.Context, userID string, targets []SubscriptionTarget, severity string) (*model.Subscription, error) {
	sev := strings.ToUpper(severity)
	sevMap := map[string]int16{
		"INFO":     1,
		"LOW":      2,
		"MEDIUM":   3,
		"HIGH":     4,
		"CRITICAL": 5,
	}
	levelID, ok := sevMap[sev]
	if !ok {
		levelID = 2 // Default to LOW
	}
	sub := &model.Subscription{
		UserID:            mustParseUUID(userID),
		SeverityThreshold: levelID,
	}
	var modelTargets []model.SubscriptionTarget
	for _, t := range targets {
		var typeID int
		switch strings.ToLower(t.TargetType) {
		case "cve_source":
			typeID = 1
		case "vendor":
			typeID = 2
		case "product":
			typeID = 3
		default:
			typeID = 1
		}
		modelTargets = append(modelTargets, model.SubscriptionTarget{
			TargetTypeID: typeID,
			TargetID:     mustParseUUID(t.TargetID),
		})
	}
	if err := s.repo.CreateSubscription(ctx, sub, modelTargets); err != nil {
		return nil, err
	}
	return sub, nil
}

// ListSubscriptions returns all subscriptions for a user.
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID string) ([]model.Subscription, error) {
	return s.repo.GetSubscriptionsByUser(ctx, userID)
}

// DeleteSubscription deletes a subscription for a user.
func (s *SubscriptionService) DeleteSubscription(ctx context.Context, userID string, subscriptionID string) error {
	return s.repo.DeleteSubscription(ctx, userID, subscriptionID)
}

// helper to parse string to uuid.UUID
func mustParseUUID(s string) uuid.UUID {
	u, _ := uuid.Parse(s)
	return u
}
