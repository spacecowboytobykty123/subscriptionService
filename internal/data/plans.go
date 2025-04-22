package data

import (
	"database/sql"
)

type Plan struct {
	ID          int32
	Name        string
	Desc        string
	RentalLimit int32
	Price       int32
	Duration    int32
}

type PlanModel struct {
	DB *sql.DB
}
