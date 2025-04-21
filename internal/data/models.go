package data

import "database/sql"

type Models struct {
	Plans PlanModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Plans: PlanModel{DB: db},
	}
}
