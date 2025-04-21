package subscription

import (
	"context"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"google.golang.org/grpc"
)

type serverAPI struct {
	subs.UnimplementedSubscriptionServer
	subs Subscription
}

type Subscription interface {
	Subscribe(ctx context.Context, userId int64, planId int32) (int64, bool)
	ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) bool
	Unsubscribe(ctx context.Context, userId int64) bool
	GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, string)
	CheckSub(ctx context.Context, userId int64) bool
}

func Register(gRPC *grpc.Server, subscription Subscription) {
	subs.RegisterSubscriptionServer(gRPC, &serverAPI{subs: subscription})
}

func (s *serverAPI) Subscribe(ctx context.Context, r *subs.SubsRequest) (*subs.SubsResponse, error) {
	panic("Implement me!")
}

func (s *serverAPI) ChangeSubPlan(ctx context.Context, r *subs.ChangePlanRequest) (*subs.ChangePlanResponse, error) {
	panic("Implement me!")
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
