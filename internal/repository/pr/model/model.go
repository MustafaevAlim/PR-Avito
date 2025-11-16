package model

import (
	"time"

	"github.com/google/uuid"
)

type PullRequest struct {
	ID                uuid.UUID `db:"id"`
	Name              string    `db:"name"`
	AuthorID          uuid.UUID `db:"author_id"`
	Status            string    `db:"status"`
	AssignedReviewers []uuid.UUID
	CreatedAt         *time.Time `db:"create_at"`
	MergedAt          *time.Time `db:"merged_at"`
}

type PullRequestShort struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	AuthorID uuid.UUID `db:"author_id"`
	Status   string    `db:"status"`
}
