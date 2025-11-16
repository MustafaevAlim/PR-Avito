package repository

import (
	"context"

	"github.com/google/uuid"

	"PR/internal/model"
)

type PullRequestRepository interface {
	CreatePR(ctx context.Context, pr *model.PullRequest) error
	CreatePRReviewers(ctx context.Context, pr *model.PullRequest) error
	Merge(ctx context.Context, id uuid.UUID) (*model.PullRequest, error)
	ReassignReviewers(ctx context.Context, prID, oldID, newID uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*model.PullRequest, error)
	GetViewers(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
	GetByReviewer(ctx context.Context, userID uuid.UUID) ([]*model.PullRequestShort, error)
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, teamName string) error
	CreateMembers(ctx context.Context, t *model.Team) error

	GetTeamIDByName(ctx context.Context, name string) (uuid.UUID, error)
	GetTeamByName(ctx context.Context, name string) (*model.Team, error)
}

type UserRepository interface {
	GetActiveByTeam(ctx context.Context, teamName string) ([]*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	SetActive(ctx context.Context, req *model.UserSetActive) error
}

type StatisticsRepository interface {
	GetReviewerStatistics(ctx context.Context) ([]*model.ReviewerStats, error)
	GetPRStatistics(ctx context.Context) ([]*model.PRStats, error)
}
