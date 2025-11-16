package pr

import (
	"context"
	"errors"
	"math/rand"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"PR/internal/model"
)

func (s *serv) Merge(ctx context.Context, id uuid.UUID) (*model.PullRequest, error) {

	pr, err := s.pullRequestRepo.Merge(ctx, id)
	if err != nil {
		log.Error().Msgf("%s.Merge error: %v", op, err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return pr, nil
}

func (s *serv) ReassignReviewers(ctx context.Context, oldID, prID uuid.UUID) (*model.PullRequest, uuid.UUID, error) {
	var pr *model.PullRequest
	var replaceBy uuid.UUID
	err := s.txManager.ReadCommited(ctx, func(ctx context.Context) error {
		var errTx error
		pr, errTx = s.pullRequestRepo.GetByID(ctx, prID)
		if errTx != nil {
			if errors.Is(errTx, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return errTx
		}

		if pr.Status == "MERGED" {
			return ErrPRMerged
		}

		user, errTx := s.userRepo.GetByID(ctx, oldID)
		if errTx != nil {
			if errors.Is(errTx, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return errTx
		}

		members, errTx := s.userRepo.GetActiveByTeam(ctx, user.TeamName)
		if errTx != nil {
			if errors.Is(errTx, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return errTx
		}

		replaceBy, errTx = selectNewReviewer(members, pr.AuthorID, pr.AssignedReviewers)
		if errTx != nil {
			return errTx
		}

		errTx = s.pullRequestRepo.ReassignReviewers(ctx, prID, oldID, replaceBy)
		if errTx != nil {
			if errors.Is(errTx, pgx.ErrNoRows) {
				return ErrNoAssigned
			}
			return errTx
		}
		pr, errTx = s.pullRequestRepo.GetByID(ctx, pr.ID)
		if errTx != nil {
			return errTx
		}
		return nil

	})

	if err != nil {
		log.Error().Msgf("%s.ReassignReviewers error: %v", op, err)
		return nil, uuid.UUID{}, err
	}
	return pr, replaceBy, err
}

func selectNewReviewer(members []*model.User, authorID uuid.UUID, reviewers []uuid.UUID) (uuid.UUID, error) {
	candidates := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		if member.ID != authorID && !slices.Contains(reviewers, member.ID) {
			candidates = append(candidates, member.ID)
		}
	}

	if len(candidates) == 0 {
		return uuid.UUID{}, ErrNoCandidate
	}

	return candidates[rand.Intn(len(candidates))], nil
}
