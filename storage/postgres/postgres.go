package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	subs "github.com/spacecowboytobykty123/subsProto/proto/gen/go/subscription"
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

func (s *Storage) Subscribe(ctx context.Context, userID int64, planID int32) (int64, subs.Status) {
	query := `
INSERT INTO subscriptions (user_id, plan_id, remaining_limit, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id`

	limit, durationMonths, err := s.getDetailsFromPlan(planID)
	expiresAt := addMonths(time.Now(), durationMonths)

	if err != nil {
		println(err.Error())
		return 0, subs.Status_STATUS_INTERNAL_ERROR
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var subId int64
	args := []any{userID, planID, limit, expiresAt}
	status := s.db.QueryRowContext(ctx, query, args...).Scan(&subId)
	if status != nil {
		switch {
		case errors.Is(status, sql.ErrNoRows):
			return 0, subs.Status_STATUS_NOT_SUBSCRIBED // TODO: change it to subID not found
		default:
			println(status.Error())
			return 0, subs.Status_STATUS_INTERNAL_ERROR
		}

	}
	println("db part")

	return subId, subs.Status_STATUS_OK

}

func (s *Storage) ExtractFromBalance(ctx context.Context, value int64, userId int64) (subs.Status, string, int64) {
	query := `UPDATE subscriptions
SET remaining_limit = remaining_limit - $1
WHERE user_id = $2 AND remaining_limit >= $1
RETURNING remaining_limit
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var remaining_limit int64

	err := s.db.QueryRowContext(ctx, query, value, userId).Scan(&remaining_limit)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return subs.Status_STATUS_INTERNAL_ERROR, "not enough money to buy!", 0
		default:
			return subs.Status_STATUS_INTERNAL_ERROR, "internal error!", 0
		}
	}
	return subs.Status_STATUS_OK, "extracting from balance was successful!", remaining_limit
}

func (s *Storage) AddToBalance(ctx context.Context, value int64, userId int64) (subs.Status, string, int64) {
	query := `UPDATE subscriptions
SET remaining_limit = remaining_limit + $1
WHERE user_id = $2 
RETURNING remaining_limit
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var remaining_limit int64

	err := s.db.QueryRowContext(ctx, query, value, userId).Scan(&remaining_limit)
	if err != nil {
		return subs.Status_STATUS_INTERNAL_ERROR, "internal error!", 0
	}
	return subs.Status_STATUS_OK, "adding to balance was successful!", remaining_limit
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

func (s *Storage) ChangeSubsPlan(ctx context.Context, userId int64, newPlanId int32) subs.Status {
	query := `
UPDATE subscriptions
SET plan_id = $1
WHERE user_id = $2
RETURNING id
`
	args := []any{newPlanId, userId}

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
SELECT plan_id, remaining_limit, expires_at FROM subscriptions
WHERE user_id = $1
`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var planId, remainingLimit int32
	var expires_at time.Time

	err := s.db.QueryRowContext(ctx, query, userId).Scan(&planId, &remainingLimit, &expires_at)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, "", 0, expires_at
		default:
			return 0, "", 0, expires_at
		}
	}
	planName, err := s.getNameFromPlan(planId)
	if err != nil {
		return 0, "", 0, expires_at
	}

	return planId, planName, remainingLimit, expires_at
}

func (s *Storage) CheckSubscription(ctx context.Context, userId int64) subs.Status {
	query := `
SELECT status FROM subscriptions
WHERE user_id = $1
`
	println(userId)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var subStatus string

	err := s.db.QueryRowContext(ctx, query, userId).Scan(&subStatus)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subs.Status_STATUS_NOT_SUBSCRIBED
		default:
			return subs.Status_STATUS_INTERNAL_ERROR
		}
	}

	return stringToSubsStatus(subStatus)

}

func (s *Storage) ListPlans(ctx context.Context) []*subs.Plan {
	query := `
SELECT id, name, description, rental_limit, price, duration_months FROM subscription_plans 
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	plans := []*subs.Plan{}

	for rows.Next() {
		var plan subs.Plan

		err := rows.Scan(
			&plan.PlanId,
			&plan.Name,
			&plan.Description,
			&plan.RentalLimit,
			&plan.Price,
			&plan.Duration,
		)

		if err != nil {
			return nil
		}

		plans = append(plans, &plan)
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

func (s *Storage) getNameFromPlan(planId int32) (string, error) {
	query := `
SELECT name FROM subscription_plans
WHERE id = $1
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var planName string

	err := s.db.QueryRowContext(ctx, query, planId).Scan(&planName)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", fmt.Errorf("%s:%d", "storage.postgres.getLimitFromId", ErrPlanNotFound)
		default:
			return "", fmt.Errorf("%s:%d", "storage.postgres.getLimitFromId", err)
		}
	}

	return planName, nil

}

func stringToSubsStatus(status string) subs.Status {
	if status == "active" {
		return subs.Status_STATUS_SUBSCRIBED
	} else {
		return subs.Status_STATUS_NOT_SUBSCRIBED
	}
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
