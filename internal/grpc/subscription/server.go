package subscription

import (
	"context"
	"fmt"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"subscriptionMService/internal/data"
	"subscriptionMService/internal/validator"
	"subscriptionMService/storage/postgres"
)

type serverAPI struct {
	subs.UnimplementedSubscriptionServer
	subs Subscription
}

type Subscription interface {
	Subscribe(ctx context.Context, userId int64, planId int32) (int64, subs.Status)
	ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status
	Unsubscribe(ctx context.Context, userId int64) subs.Status
	GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, string)
	CheckSub(ctx context.Context, userId int64) subs.Status
}

func Register(gRPC *grpc.Server, subscription Subscription) {
	subs.RegisterSubscriptionServer(gRPC, &serverAPI{subs: subscription})
}

func (s *serverAPI) Subscribe(ctx context.Context, r *subs.SubsRequest) (*subs.SubsResponse, error) {
	v := validator.New()

	UserID := r.GetUserId()
	PlanID := r.GetPlanId()

	subscription := data.Subscription{
		UserID: UserID,
		PlanID: PlanID,
	}

	if postgres.ValidateSubscription(v, &subscription); !v.Valid() {
		return nil, collectErrors(v)
	}

	subID, isCompleted := s.subs.Subscribe(ctx, UserID, PlanID)
	if isCompleted != subs.Status_STATUS_OK {
		return nil, mapStatusToError(isCompleted)
	}
	return &subs.SubsResponse{
		SubId:  subID,
		Status: isCompleted,
	}, nil
}

func (s *serverAPI) ChangeSubPlan(ctx context.Context, r *subs.ChangePlanRequest) (*subs.ChangePlanResponse, error) {
	v := validator.New()

	UserID := r.GetUserId()
	NewPlanID := r.GetNewPlanId()

	if postgres.ValidateSubChange(v, UserID, NewPlanID); !v.Valid() {
		return nil, collectErrors(v)
	}

	resStatus := s.subs.ChangeSubPlan(ctx, UserID, NewPlanID)

	if resStatus != subs.Status_STATUS_OK {
		return nil, mapStatusToError(resStatus)
	}

	return &subs.ChangePlanResponse{Status: resStatus}, nil

}

func (s *serverAPI) Unsubscribe(ctx context.Context, r *subs.UnSubsRequest) (*subs.UnSubsResponse, error) {
	panic("Implement me!")
}

func (s *serverAPI) GetSubDetails(ctx context.Context, r *subs.GetSubRequest) (*subs.GetSubResponse, error) {
	panic("Implement me!")
}

func (s *serverAPI) CheckSub(ctx context.Context, r *subs.CheckSubsRequest) (*subs.CheckSubsResponse, error) {
	panic("Implement me!")
}

func collectErrors(v *validator.Validator) error {
	var b strings.Builder
	for field, msg := range v.Errors {
		fmt.Fprintf(&b, "%s:%s; ", field, msg)
	}
	return status.Error(codes.InvalidArgument, b.String())
}

func mapStatusToError(code subs.Status) error {
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
