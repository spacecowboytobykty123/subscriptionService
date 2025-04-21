package data

import (
	"database/sql"
	"time"
)

type Plan struct {
	ID          int32
	Name        string
	RentalLimit int32
	Price       int32
	Duration    time.Duration
}

type PlanModel struct {
	DB *sql.DB
}
