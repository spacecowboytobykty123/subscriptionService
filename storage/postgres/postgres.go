package postgres

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"subscriptionMService/internal/data"
	"subscriptionMService/internal/validator"
	"time"
)

type Storage struct {
	db *sql.DB
}

const (
	emptyValue = 0
)

type StorageDetails struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func OpenDB(details StorageDetails) (*Storage, error) {
	db, err := sql.Open("postgres", details.DSN)

	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(details.MaxOpenConns)
	db.SetMaxIdleConns(details.MaxOpenConns)

	duration, err := time.ParseDuration(details.MaxIdleTime)

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func ValidateSubscription(v *validator.Validator, subscription *data.Subscription) {
	v.Check(subscription.UserID == emptyValue, "text", "User ID is required!")
	v.Check(subscription.PlanID == emptyValue, "text", "Plan ID is required!")
}

func ValidateSubChange(v *validator.Validator, userID int64, newPlanID int32) {
	v.Check(userID == emptyValue, "text", "User ID is required!")
	v.Check(newPlanID == emptyValue, "text", "New Plan ID is required!")
}

func (s *Storage) Subscribe(ctx context.Context, userID int64, planID int32) (int64, subs.Status) {
	panic("Do me! Storage Part")
}

func (s *Storage) Unsubscribe(ctx context.Context, userID int64) subs.Status {
	panic("Do me! Storage Part")
}

func (s *Storage) ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, time.Time) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) CheckSub(ctx context.Context, userId int64) subs.Status {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) ListPlans(ctx context.Context) []*data.Plan {
	//TODO implement me
	panic("implement me")
}
