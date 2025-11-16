package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `db:"id"`
	Username string    `db:"username"`
	TeamName string    `db:"team_name"`
	IsActive bool      `db:"is_active"`
}
