package data

import "time"

type Subscription struct {
	ID             int64
	UserID         int64
	PlanID         int32
	RemainingLimit int32
	ExpiresAt      time.Time
}
