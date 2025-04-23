package subscription

import (
	"context"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"subscriptionMService/internal/jsonlog"
	"time"
)

type Subscription struct {
	log          *jsonlog.Logger
	subProvider  subProvider
	planProvider planProvider
	tokenTTL     time.Duration
}

type subProvider interface {
	Subscribe(ctx context.Context, userId int64, planId int32) (int64, subs.Status)
	ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status
	Unsubscribe(ctx context.Context, userId int64) subs.Status
	GetSubDetails(ctx context.Context, userId int64) (int32, string, int32, time.Time)
	CheckSub(ctx context.Context, userId int64) subs.Status
}

type planProvider interface {
	ListPlans(ctx context.Context) []subs.Plan
}

func New(
	log *jsonlog.Logger,
	subProvider subProvider,
	planProvider planProvider,
	tokenTTL time.Duration,
) *Subscription {
	return &Subscription{
		log:          log,
		subProvider:  subProvider,
		planProvider: planProvider,
		tokenTTL:     tokenTTL,
	}
}

func (s *Subscription) Subscribe(ctx context.Context, userId int64, planId int32) (int64, subs.Status) {
	s.log.PrintInfo("Attempting to subscribe user", nil)

	isSubscribed := s.subProvider.CheckSub(ctx, userId)
	if isSubscribed == subs.Status_STATUS_SUBSCRIBED {
		return 0, subs.Status_STATUS_ALREADY_SUBSCRIBED
	}

	subId, subStatus := s.subProvider.Subscribe(ctx, userId, planId)

	if subStatus != subs.Status_STATUS_OK {
		s.log.PrintError(s.MapStatusToError(subStatus), map[string]string{
			"method": "server.Subscribe",
		})
		return 0, subStatus
	}

	return subId, subStatus
}

func (s *Subscription) ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status {
	s.log.PrintInfo("Attempting change subscription plan", nil)

	isCompleted := s.subProvider.ChangeSubPlan(ctx, userId, newPlanId)

	return isCompleted
}

func (s *Subscription) Unsubscribe(ctx context.Context, userId int64) subs.Status {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) GetSubDetails(ctx context.Context, userId int64) (int32, string, int32, string) {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) CheckSub(ctx context.Context, userId int64) subs.Status {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) ListPlans(ctx context.Context) []subs.Plan {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) MapStatusToError(code subs.Status) error {
	switch code {
	case subs.Status_STATUS_INVALID_PLAN:
		return status.Error(codes.InvalidArgument, "Invalid plan")
	case subs.Status_STATUS_INVALID_USER:
		return status.Error(codes.InvalidArgument, "Invalid user")
	case subs.Status_STATUS_ALREADY_SUBSCRIBED:
		return status.Error(codes.FailedPrecondition, "User already subscribed")
	case subs.Status_STATUS_SUBSCRIPTION_NOTFOUND:
		return status.Error(codes.NotFound, "Subscription not found")
	default:
		return status.Error(codes.Internal, "Internal error")
	}
}
