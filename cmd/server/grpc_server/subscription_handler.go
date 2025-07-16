package grpc_server

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	secra_v1 "gitlab.com/jacky850509/secra/api/gen/v1"
	"gitlab.com/jacky850509/secra/internal/service"
)

// SubscriptionHandler implements secra_v1.SubscriptionServiceServer.
type SubscriptionHandler struct {
	subService *service.SubscriptionService
	secra_v1.UnimplementedSubscriptionServiceServer
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(subSvc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subSvc}
}

// CreateSubscription creates a new subscription.
func (h *SubscriptionHandler) CreateSubscription(ctx context.Context, req *secra_v1.CreateSubscriptionRequest) (*secra_v1.CreateSubscriptionResponse, error) {
	targets := make([]service.SubscriptionTarget, len(req.Targets))
	for i, t := range req.Targets {
		targets[i] = service.SubscriptionTarget{
			TargetType: t.TargetType,
			TargetID:   t.TargetId,
		}
	}
	sub, err := h.subService.CreateSubscription(ctx, req.UserId, targets, req.SeverityThreshold)
	if err != nil {
		return nil, err
	}
	resp := &secra_v1.CreateSubscriptionResponse{
		Subscription: &secra_v1.Subscription{
			Id:                sub.ID.String(),
			UserId:            sub.UserID.String(),
			SeverityThreshold: service.SeverityToString(sub.SeverityThreshold),
			CreatedAt:         sub.CreatedAt.Format(time.RFC3339),
		},
	}
	for _, t := range sub.Targets {
		resp.Subscription.Targets = append(resp.Subscription.Targets, &secra_v1.SubscriptionTarget{
			TargetType: service.SeverityToString(int16(t.TargetTypeID)),
			TargetId:   t.TargetID.String(),
		})
	}
	return resp, nil
}

// ListSubscriptions lists all subscriptions for a user.
func (h *SubscriptionHandler) ListSubscriptions(ctx context.Context, req *secra_v1.ListSubscriptionsRequest) (*secra_v1.ListSubscriptionsResponse, error) {
	subs, err := h.subService.ListSubscriptions(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	resp := &secra_v1.ListSubscriptionsResponse{Total: int32(len(subs))}
	for _, s := range subs {
		item := &secra_v1.Subscription{
			Id:                s.ID.String(),
			UserId:            s.UserID.String(),
			SeverityThreshold: service.SeverityToString(s.SeverityThreshold),
			CreatedAt:         s.CreatedAt.Format(time.RFC3339),
		}
		for _, t := range s.Targets {
			item.Targets = append(item.Targets, &secra_v1.SubscriptionTarget{
				TargetType: service.SeverityToString(int16(t.TargetTypeID)),
				TargetId:   t.TargetID.String(),
			})
		}
		resp.Subscriptions = append(resp.Subscriptions, item)
	}
	return resp, nil
}

// DeleteSubscription deletes a subscription.
func (h *SubscriptionHandler) DeleteSubscription(ctx context.Context, req *secra_v1.DeleteSubscriptionRequest) (*emptypb.Empty, error) {
	if err := h.subService.DeleteSubscription(ctx /* userID from context? */, "", req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
