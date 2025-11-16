package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	TeamName string    `json:"team_name"`
	IsActive bool      `json:"is_active"`
}

type UserSetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserSetActive struct {
	UserID   uuid.UUID
	IsActive bool
}
