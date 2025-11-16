package model

import "github.com/google/uuid"

type TeamMember struct {
	ID       uuid.UUID `db:"id"`
	TeamName string    `db:"team_name"`
	Username string    `db:"username"`
	IsActive bool      `db:"is_active"`
}
