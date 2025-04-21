package subscription

import (
	"context"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"subscriptionMService/internal/data"
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
	GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, time.Time)
	CheckSub(ctx context.Context, userId int64) subs.Status
}

type planProvider interface {
	ListPlans(ctx context.Context) []*data.Plan
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
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) Unsubscribe(ctx context.Context, userId int64) subs.Status {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, string) {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) CheckSub(ctx context.Context, userId int64) subs.Status {
	//TODO implement me
	panic("implement me")
}
