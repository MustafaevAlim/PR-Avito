package pr

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"

	"PR/internal/model"
)

func (s *serv) Create(ctx context.Context, p *model.PullRequestShort) (*model.PullRequest, error) {

	var pr *model.PullRequest
	err := s.txManager.ReadCommited(ctx, func(ctx context.Context) error {
		var errTx error

		author, errTx := s.userRepo.GetByID(ctx, p.AuthorID)
		if errTx != nil {
			if errors.Is(errTx, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return errTx
		}

		if !author.IsActive {
			return ErrNotActive
		}

		teamMembers, errTx := s.userRepo.GetActiveByTeam(ctx, author.TeamName)
		if errTx != nil {
			return errTx
		}

		reviewers := selectReviewers(teamMembers, p.AuthorID, 2)

		pr = &model.PullRequest{
			ID:                p.ID,
			Name:              p.Name,
			AuthorID:          p.AuthorID,
			Status:            "OPEN",
			AssignedReviewers: reviewers,
		}

		errTx = s.pullRequestRepo.CreatePR(ctx, pr)
		if errTx != nil {
			var pgErr *pgconn.PgError
			if errors.As(errTx, &pgErr) {
				if pgErr.Code == "23505" {
					return ErrPRExists
				}
			}
			return errTx
		}

		errTx = s.pullRequestRepo.CreatePRReviewers(ctx, pr)
		if errTx != nil {
			return errTx
		}
		return nil
	})

	if err != nil {
		log.Error().Msgf("%s.Create error: %v", op, err)
		return nil, err
	}
	return pr, nil
}

func selectReviewers(members []*model.User, authorID uuid.UUID, maxCount int) []uuid.UUID {
	reviewers := make([]uuid.UUID, 0, maxCount)
	for _, m := range members {
		if m.ID != authorID && len(reviewers) < maxCount {
			reviewers = append(reviewers, m.ID)
		}
	}
	return reviewers
}
