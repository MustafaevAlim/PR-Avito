package model

import "github.com/google/uuid"

type PRStats struct {
	PRID          uuid.UUID `json:"pr_id" db:"pr_id"`
	PRName        string    `json:"pr_name" db:"pr_name"`
	ReviewerCount int       `json:"reviewer_count" db:"reviewer_count"`
	Status        string    `json:"status" db:"status"`
}

type ReviewerStats struct {
	ReviewerID    uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	ReviewerName  string    `json:"reviewer_name" db:"reviewer_name"`
	AssignedCount int       `json:"assigned_count" db:"assigned_count"`
}
