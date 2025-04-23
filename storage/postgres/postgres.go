package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func ValidateUser(v *validator.Validator, userID int64) {
	v.Check(userID == emptyValue, "text", "User ID is required!")
}

func (s *Storage) Subscribe(ctx context.Context, userID int64, planID int32) (int64, subs.Status) {
	query := `
INSERT INTO subscriptions (user_id, plan_id, remaining_limit, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id`

	limit, durationMonths, err := s.getDetailsFromPlan(planID)
	expiresAt := addMonths(time.Now(), durationMonths)

	if err != nil {
		return 0, subs.Status_STATUS_INTERNAL_ERROR
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var subId int64
	args := []any{userID, planID, limit, expiresAt}
	status := s.db.QueryRowContext(ctx, query, args).Scan(&subId)
	if status != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, subs.Status_STATUS_NOT_SUBSCRIBED // TODO: change it to subID not found
		default:
			return 0, subs.Status_STATUS_INTERNAL_ERROR
		}

	}

	return subId, subs.Status_STATUS_OK

}

func (s *Storage) Unsubscribe(ctx context.Context, userID int64) subs.Status {
	query := `
DELETE FROM subscriptions
WHERE user_id = $1
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return subs.Status_STATUS_INTERNAL_ERROR
	}
	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return subs.Status_STATUS_INTERNAL_ERROR
	}

	if rowsAffected == 0 {
		return subs.Status_STATUS_NOT_SUBSCRIBED
	}
	return subs.Status_STATUS_OK
}

func (s *Storage) ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status {
	query := `
UPDATE subscriptions
SET plan_id = $1
WHERE user_id = 2$
RETURNING id
`
	args := []any{userId, newPlanId}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var subId int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&subId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subs.Status_STATUS_INVALID_USER // TODO: мб поменять
		default:
			return subs.Status_STATUS_INTERNAL_ERROR
		}
	}
	return subs.Status_STATUS_OK
}

func (s *Storage) GetSubDetails(ctx context.Context, userId int64) (int32, string, int32, time.Time) {
	query := `
SELECT 
`
}

func (s *Storage) CheckSub(ctx context.Context, userId int64) subs.Status {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) ListPlans(ctx context.Context) []subs.Plan {
	query := `
SELECT name, desc, rental_limit, price, duration_months FROM subscription_plans 
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	plans := []subs.Plan{}

	for rows.Next() {
		var plan subs.Plan

		err := rows.Scan(
			plan.PlanId,
			plan.Name,
			plan.Description,
			plan.RentalLimit,
			plan.Price,
			plan.Duration,
		)

		if err != nil {
			return nil
		}

		plans = append(plans, plan)
	}

	if err = rows.Err(); err != nil {
		return nil
	}

	return plans

}

func (s *Storage) getDetailsFromPlan(planId int32) (int32, subs.Duration, error) {
	query := `
SELECT rental_limit, duration_months FROM subscription_plans
WHERE id = $1
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var rentalLimit int32
	var durationsMonths subs.Duration

	err := s.db.QueryRowContext(ctx, query, planId).Scan(&rentalLimit, &durationsMonths)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, subs.Duration(0), fmt.Errorf("%s:%d", "storage.postgres.getLimitFromId", ErrPlanNotFound)
		default:
			return 0, subs.Duration(0), fmt.Errorf("%s:%d", "storage.postgres.getLimitFromId", err)
		}
	}

	return rentalLimit, durationsMonths, nil

}

func addMonths(t time.Time, months subs.Duration) time.Time {
	year := t.Year()
	month := t.Month()
	day := t.Day()

	month += time.Month(months)
	newTime := time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	for newTime.Month() != month%12 {
		day--
		newTime = time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	}

	return newTime
}
