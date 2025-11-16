package service

import (
	"context"

	"github.com/google/uuid"

	"PR/internal/model"
)

type PullRequestService interface {
	Create(ctx context.Context, p *model.PullRequestShort) (*model.PullRequest, error)
	Merge(ctx context.Context, id uuid.UUID) (*model.PullRequest, error)
	ReassignReviewers(ctx context.Context, oldID, prID uuid.UUID) (*model.PullRequest, uuid.UUID, error)

	GetByReviewer(ctx context.Context, userID uuid.UUID) ([]*model.PullRequestShort, error)
}

type TeamService interface {
	Create(ctx context.Context, t *model.Team) error

	GetTeamByName(ctx context.Context, name string) (*model.Team, error)
}

type UserService interface {
	SetActive(ctx context.Context, req *model.UserSetActive) (*model.User, error)
}

type StatisticsService interface {
	GetReviewerStatistics(ctx context.Context) ([]*model.ReviewerStats, error)
	GetPRStatistics(ctx context.Context) ([]*model.PRStats, error)
}
