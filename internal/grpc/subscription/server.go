package subscription

import (
	"context"
	"fmt"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"subscriptionMService/internal/contextkeys"
	"subscriptionMService/internal/validator"
)

type serverAPI struct {
	subs.UnimplementedSubscriptionServer
	subs Subscription
}

type Subscription interface {
	Subscribe(ctx context.Context, planId int32) (int64, subs.Status)
	ChangeSubsPlan(ctx context.Context, newPlanId int32) subs.Status
	Unsubscribe(ctx context.Context) subs.Status
	GetSubDetails(ctx context.Context) (int32, string, int32, string)
	CheckSubscription(ctx context.Context) subs.Status
	ListPlans(ctx context.Context) []*subs.Plan
}

func Register(gRPC *grpc.Server, subscription Subscription) {
	subs.RegisterSubscriptionServer(gRPC, &serverAPI{subs: subscription})
}

func (s *serverAPI) Subscribe(ctx context.Context, r *subs.SubsRequest) (*subs.SubsResponse, error) {
	v := validator.New()

	planID := r.GetPlanId()

	v.Check(planID != 0, "text", "plan_id not found or invalid")

	if !v.Valid() {
		return nil, collectErrors(v)
	}

	subID, isCompleted := s.subs.Subscribe(ctx, planID)

	// TODO: Отправить сообщение на почту

	return &subs.SubsResponse{
		SubId:  subID,
		Status: isCompleted,
	}, nil
}

func (s *serverAPI) ChangeSubsPlan(ctx context.Context, r *subs.ChangePlanRequest) (*subs.ChangePlanResponse, error) {
	v := validator.New()

	NewPlanID := r.GetNewPlanId()

	v.Check(NewPlanID != 0, "text", "plan_id not found or invalid")

	if !v.Valid() {
		return nil, collectErrors(v)
	}

	resStatus := s.subs.ChangeSubsPlan(ctx, NewPlanID)

	if resStatus != subs.Status_STATUS_OK {
		return nil, s.MapStatusToError(resStatus)
	}

	return &subs.ChangePlanResponse{Status: resStatus}, nil

}

func (s *serverAPI) Unsubscribe(ctx context.Context, r *subs.UnSubsRequest) (*subs.UnSubsResponse, error) {
	resStatus := s.subs.Unsubscribe(ctx)
	if resStatus != subs.Status_STATUS_OK {
		return nil, s.MapStatusToError(resStatus)
	}

	return &subs.UnSubsResponse{Status: resStatus}, nil

}

func (s *serverAPI) GetSubDetails(ctx context.Context, r *subs.GetSubRequest) (*subs.GetSubResponse, error) {
	userIDRaw := ctx.Value(contextkeys.UserIDKey)
	userID, ok := userIDRaw.(int64)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user_id not found or invalid")
	}
	planId, planName, remainingLimit, expiresAt := s.subs.GetSubDetails(ctx)
	return &subs.GetSubResponse{
		UserId:         userID,
		PlanId:         planId,
		PlanName:       planName,
		RemainingLimit: remainingLimit,
		ExpiresAt:      expiresAt,
	}, nil
}

func (s *serverAPI) CheckSubscription(ctx context.Context, r *subs.CheckSubsRequest) (*subs.CheckSubsResponse, error) {
	isSubscribed := s.subs.CheckSubscription(ctx)
	return &subs.CheckSubsResponse{SubStatus: isSubscribed}, nil
}

func (s *serverAPI) ListPlans(ctx context.Context, r *subs.PlansRequest) (*subs.PlansResponse, error) {
	plans := s.subs.ListPlans(ctx)
	planPointers := make([]*subs.Plan, len(plans))
	for i := range plans {
		planPointers[i] = plans[i]
	}
	return &subs.PlansResponse{Plans: planPointers}, nil
}

func collectErrors(v *validator.Validator) error {
	var b strings.Builder
	for field, msg := range v.Errors {
		fmt.Fprintf(&b, "%s:%s; ", field, msg)
	}
	return status.Error(codes.InvalidArgument, b.String())
}

func (s *serverAPI) MapStatusToError(code subs.Status) error {
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
