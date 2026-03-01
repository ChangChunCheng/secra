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
	UpdateThreshold(ctx context.Context, userID string, subscriptionID string, severity string) error
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
		m = make(map[int]string)
	}
	return &SubscriptionService{
		repo:          r,
		targetTypeMap: m,
	}
}

type SubscriptionTarget struct {
	TargetType string
	TargetID   string
}

func (s *SubscriptionService) SeverityToString(level int16) string {
	sevMap := map[int16]string{
		1: "INFO",
		2: "LOW",
		3: "MEDIUM",
		4: "HIGH",
		5: "CRITICAL",
	}
	if name, ok := sevMap[level]; ok {
		return name
	}
	return "UNKNOWN"
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, userID string, targets []SubscriptionTarget, severity string) (*model.Subscription, error) {
	levelID := s.severityToLevel(severity)
	sub := &model.Subscription{
		UserID:            mustParseUUID(userID),
		SeverityThreshold: levelID,
	}
	var modelTargets []model.SubscriptionTarget
	for _, t := range targets {
		var typeID int
		switch strings.ToLower(t.TargetType) {
		case "cve_source": typeID = 1
		case "vendor":     typeID = 2
		case "product":    typeID = 3
		default:           typeID = 1
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

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID string) ([]model.Subscription, error) {
	return s.repo.GetSubscriptionsByUser(ctx, userID)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, userID string, subscriptionID string) error {
	return s.repo.DeleteSubscription(ctx, userID, subscriptionID)
}

func (s *SubscriptionService) UpdateThreshold(ctx context.Context, userID string, subscriptionID string, severity string) error {
	levelID := s.severityToLevel(severity)
	return s.repo.UpdateThreshold(ctx, userID, subscriptionID, levelID)
}

func (s *SubscriptionService) severityToLevel(severity string) int16 {
	sev := strings.ToUpper(severity)
	sevMap := map[string]int16{
		"INFO":     1,
		"LOW":      2,
		"MEDIUM":   3,
		"HIGH":     4,
		"CRITICAL": 5,
	}
	if levelID, ok := sevMap[sev]; ok {
		return levelID
	}
	return 3 // Default to MEDIUM
}

func mustParseUUID(s string) uuid.UUID {
	u, _ := uuid.Parse(s)
	return u
}
