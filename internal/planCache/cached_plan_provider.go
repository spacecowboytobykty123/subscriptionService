package planCache

import (
	"context"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"sync"
	"time"
)

type PlanProvider interface {
	ListPlans(ctx context.Context) []*subs.Plan
}

type CachedPlanProvider struct {
	underlying PlanProvider
	cache      []*subs.Plan
	cacheTime  time.Time
	ttl        time.Duration
	mu         sync.Mutex
}

func NewCachedPlanProvider(planProvider PlanProvider, ttl time.Duration) *CachedPlanProvider {
	return &CachedPlanProvider{
		underlying: planProvider,
		ttl:        ttl,
	}
}

func (c *CachedPlanProvider) ListPlans(ctx context.Context) []*subs.Plan {
	println("listplanCache")
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.cacheTime) > c.ttl && c.cache != nil {
		return c.cache
	}

	plans := c.underlying.ListPlans(ctx)
	if plans != nil {
		c.cache = plans
	}
	return plans

}
