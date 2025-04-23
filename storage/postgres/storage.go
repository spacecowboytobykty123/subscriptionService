package postgres

import "errors"

var (
	ErrUserSubscribed = errors.New("user already subscribed")
	ErrPlanNotFound   = errors.New("plan not found")
	ErrSubNotFound    = errors.New("subscription not found")
)
