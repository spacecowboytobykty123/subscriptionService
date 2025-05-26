package subscription

import (
	"context"
	"errors"
	"fmt"
	bckt "github.com/spacecowboytobykty123/bucketProto/gen/go/bucket"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	bcktgrpc "subscriptionMService/internal/clients/bucket/grpc"
	"subscriptionMService/internal/contextkeys"
	"subscriptionMService/internal/jsonlog"
	"subscriptionMService/internal/planCache"
	"time"
)

type Subscription struct {
	log           *jsonlog.Logger
	subProvider   subProvider
	planProvider  planCache.PlanProvider
	bucketService *bcktgrpc.BucketClient
	tokenTTL      time.Duration
}

type subProvider interface {
	Subscribe(ctx context.Context, userId int64, planId int32) (int64, subs.Status)
	ChangeSubsPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status
	Unsubscribe(ctx context.Context, userId int64) subs.Status
	GetSubDetails(ctx context.Context, userId int64) (int32, string, int32, time.Time)
	CheckSubscription(ctx context.Context, userId int64) subs.Status
	ExtractFromBalance(ctx context.Context, value int64, userId int64) (subs.Status, string, int64)
	AddToBalance(ctx context.Context, value int64, userId int64) (subs.Status, string, int64)
}

//type planProvider interface {
//	ListPlans(ctx context.Context) []*subs.Plan
//}

func New(
	log *jsonlog.Logger,
	subProvider subProvider,
	planProvider planCache.PlanProvider,
	bucketService *bcktgrpc.BucketClient,
	tokenTTL time.Duration,
) *Subscription {
	return &Subscription{
		log:           log,
		subProvider:   subProvider,
		planProvider:  planProvider,
		bucketService: bucketService,
		tokenTTL:      tokenTTL,
	}
}

func (s *Subscription) ExtractFromBalance(ctx context.Context, value int64) (subs.Status, string, int64) {
	userId, err := getUserFromContext(ctx)
	if err != nil {
		return subs.Status_STATUS_INVALID_USER, "invalid user!", 0
	}
	isSubscribed := s.subProvider.CheckSubscription(ctx, userId)
	if isSubscribed == subs.Status_STATUS_NOT_SUBSCRIBED {
		return subs.Status_STATUS_NOT_SUBSCRIBED, "user is not subscribed", 0
	}

	opStatus, msg, valueLeft := s.subProvider.ExtractFromBalance(ctx, value, userId)
	return opStatus, msg, valueLeft
}

func (s *Subscription) AddToBalance(ctx context.Context, value int64) (subs.Status, string, int64) {
	userId, err := getUserFromContext(ctx)
	if err != nil {
		return subs.Status_STATUS_INVALID_USER, "invalid user!", 0
	}
	isSubscribed := s.subProvider.CheckSubscription(ctx, userId)
	if isSubscribed == subs.Status_STATUS_NOT_SUBSCRIBED {
		return subs.Status_STATUS_INVALID_USER, "invalid user!", 0
	}
	opStatus, msg, valueLeft := s.subProvider.AddToBalance(ctx, value, userId)
	return opStatus, msg, valueLeft

}

func (s *Subscription) Subscribe(ctx context.Context, planId int32) (int64, subs.Status) {
	s.log.PrintInfo("Attempting to subscribe user", nil)
	userId, err := getUserFromContext(ctx)
	if err != nil {
		return 0, subs.Status_STATUS_INVALID_USER
	}

	isSubscribed := s.subProvider.CheckSubscription(ctx, userId)
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
	bucketResp := s.bucketService.CreateBucket(ctx)
	if bucketResp.Status != bckt.OperationStatus_STATUS_OK {
		s.log.PrintError(fmt.Errorf("could not create bucket for new user"), map[string]string{
			"method": "server.subscribe",
		})
		return 0, subs.Status_STATUS_INTERNAL_ERROR
	}

	return subId, subStatus
}

func (s *Subscription) ChangeSubsPlan(ctx context.Context, newPlanId int32) subs.Status {
	s.log.PrintInfo("Attempting change subscription plan", nil)
	userId, err := getUserFromContext(ctx)
	if err != nil {
		return subs.Status_STATUS_INVALID_USER
	}

	isCompleted := s.subProvider.ChangeSubsPlan(ctx, userId, newPlanId)

	return isCompleted
}

func (s *Subscription) Unsubscribe(ctx context.Context) subs.Status {
	s.log.PrintInfo("Attempting to unsubscribe user", nil)
	userId, err := getUserFromContext(ctx)
	if err != nil {
		return subs.Status_STATUS_INVALID_USER
	}

	isSubscribed := s.subProvider.CheckSubscription(ctx, userId)
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

func (s *Subscription) GetSubDetails(ctx context.Context) (int32, string, int32, string) {

	userId, err := getUserFromContext(ctx)
	if err != nil {
		return 0, "", 0, ""
	}
	isSubscribed := s.subProvider.CheckSubscription(ctx, userId)
	if isSubscribed == subs.Status_STATUS_NOT_SUBSCRIBED {
		return 0, "", 0, ""
	}
	s.log.PrintInfo("Fetching subscription details", map[string]string{
		"userId": fmt.Sprint(userId),
	})

	planId, planName, remainingLimit, expiresAt := s.subProvider.GetSubDetails(ctx, userId)
	return planId, planName, remainingLimit, expiresAt.Format(time.RFC3339)
}

func (s *Subscription) CheckSubscription(ctx context.Context) subs.Status {
	s.log.PrintInfo("Checking if user is subscribed", map[string]string{})

	userId, err := getUserFromContext(ctx)
	if err != nil {
		return subs.Status_STATUS_INVALID_USER
	}
	s.log.PrintInfo("Checking if user is subscribed", map[string]string{
		"userId": fmt.Sprint(userId),
	})

	isSubscribed := s.subProvider.CheckSubscription(ctx, userId)
	if isSubscribed != subs.Status_STATUS_SUBSCRIBED {
		s.log.PrintInfo("User not subscribed", nil)
	}
	return isSubscribed
}

func (s *Subscription) ListPlans(ctx context.Context) []*subs.Plan {
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

func getUserFromContext(ctx context.Context) (int64, error) {
	println("getUserFromContext")
	val := ctx.Value(contextkeys.UserIDKey)
	userID, ok := val.(int64)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "user id is missing or invalid in context")
	}

	return userID, nil

}
