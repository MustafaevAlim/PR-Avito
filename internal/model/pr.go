package model

import (
	"time"

	"github.com/google/uuid"
)

type PullRequestInReassign struct {
	PrID          string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

type PullRequestInMerge struct {
	ID string `json:"pull_request_id"`
}

type PullRequestInCreate struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type PullRequest struct {
	ID                uuid.UUID   `json:"pull_request_id"`
	Name              string      `json:"pull_request_name"`
	AuthorID          uuid.UUID   `json:"author_id"`
	Status            string      `json:"status"`
	AssignedReviewers []uuid.UUID `json:"assigned_reviewers"`
	CreatedAt         *time.Time  `json:"createdAt"`
	MergedAt          *time.Time  `json:"mergedAt"`
}

type PullRequestShort struct {
	ID       uuid.UUID `json:"pull_request_id"`
	Name     string    `json:"pull_request_name"`
	AuthorID uuid.UUID `json:"author_id"`
	Status   string    `json:"status"`
}
