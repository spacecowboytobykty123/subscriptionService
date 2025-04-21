package postgres

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"subscriptionMService/internal/data"
	"time"
)

type Storage struct {
	db *sql.DB
}

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

func (s *Storage) Subscribe(ctx context.Context, userID int64, planID int32) (int64, bool) {
	panic("Do me! Storage Part")
}

func (s *Storage) Unsubscribe(ctx context.Context, userID int64) bool {
	panic("Do me! Storage Part")
}

func (s *Storage) ChangeSubPlan(ctx context.Context, userId int64, newPlanId int32) bool {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) GetSubDetails(ctx context.Context, userId int64) (int64, int32, string, int32, time.Time) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) CheckSub(ctx context.Context, userId int64) bool {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) ListPlans(ctx context.Context) []*data.Plan {
	//TODO implement me
	panic("implement me")
}
