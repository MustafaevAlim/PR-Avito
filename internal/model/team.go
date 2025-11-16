package model

import "github.com/google/uuid"

type CreateTeamRequest struct {
	TeamName string          `json:"team_name"`
	Members  []MemberRequest `json:"members"`
}

type MemberRequest struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamMember struct {
	ID       uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	IsActive bool      `json:"is_active"`
}

type Team struct {
	TeamName string        `json:"team_name"`
	Members  []*TeamMember `json:"members"`
}
