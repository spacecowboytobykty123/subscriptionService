package subscription

import (
	"context"
	"errors"
	"fmt"
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
	ChangeSubsPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status
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

func (s *Subscription) ChangeSubsPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status {
	s.log.PrintInfo("Attempting change subscription plan", nil)

	isCompleted := s.subProvider.ChangeSubsPlan(ctx, userId, newPlanId)

	return isCompleted
}

func (s *Subscription) Unsubscribe(ctx context.Context, userId int64) subs.Status {
	s.log.PrintInfo("Attempting to unsubscribe user", nil)

	isSubscribed := s.subProvider.CheckSub(ctx, userId)
	if isSubscribed == subs.Status_STATUS_NOT_SUBSCRIBED {
		return subs.Status_STATUS_NOT_SUBSCRIBED
	}

	// TODO: Какое то сообщение о скидке(чтобы оставить клиента)

	isCompleted := s.subProvider.Unsubscribe(ctx, userId)

	if isCompleted != subs.Status_STATUS_OK {
		s.log.PrintError(s.MapStatusToError(isCompleted), map[string]string{
			"method": "server.Subscribe",
		})
		return isCompleted
	}
	return isCompleted

}

func (s *Subscription) GetSubDetails(ctx context.Context, userId int64) (int32, string, int32, string) {

	isSubscribed := s.subProvider.CheckSub(ctx, userId)
	if isSubscribed == subs.Status_STATUS_NOT_SUBSCRIBED {
		return 0, "", 0, ""
	}
	s.log.PrintInfo("Fetching subscription details", map[string]string{
		"userId": fmt.Sprint(userId),
	})

	planId, planName, remainingLimit, expiresAt := s.subProvider.GetSubDetails(ctx, userId)
	return planId, planName, remainingLimit, expiresAt.Format(time.RFC3339)
}

func (s *Subscription) CheckSub(ctx context.Context, userId int64) subs.Status {
	s.log.PrintInfo("Checking if user is subscribed", map[string]string{
		"userId": fmt.Sprint(userId),
	})

	isSubscribed := s.subProvider.CheckSub(ctx, userId)
	if isSubscribed != subs.Status_STATUS_SUBSCRIBED {
		s.log.PrintInfo("User not subscribed", nil)
	}
	return isSubscribed
}

func (s *Subscription) ListPlans(ctx context.Context) []subs.Plan {
	s.log.PrintInfo("Listing available plans", nil)

	plans := s.planProvider.ListPlans(ctx)
	if plans == nil {
		s.log.PrintError(errors.New("failed to fetch plans"), nil)
	}
	return plans
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
