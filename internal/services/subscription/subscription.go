package subscription

import (
	"context"
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
	Subscribe(ctx context.Context, userId int64, planId int32) (int64, bool)
	ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) bool
	Unsubscribe(ctx context.Context, userId int64) bool
	GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, time.Time)
	CheckSub(ctx context.Context, userId int64) bool
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

func (s *Subscription) Subscribe(ctx context.Context, userId int64, planId int32) (int64, bool) {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) bool {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) Unsubscribe(ctx context.Context, userId int64) bool {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, string) {
	//TODO implement me
	panic("implement me")
}

func (s *Subscription) CheckSub(ctx context.Context, userId int64) bool {
	//TODO implement me
	panic("implement me")
}
