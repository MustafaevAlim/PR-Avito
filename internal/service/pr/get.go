package pr

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"PR/internal/model"
)

func (s *serv) GetByReviewer(ctx context.Context, userID uuid.UUID) ([]*model.PullRequestShort, error) {
	prs, err := s.pullRequestRepo.GetByReviewer(ctx, userID)
	if err != nil {
		log.Error().Msgf("%s.GetByReviewer error: %v", op, err)
		return nil, err
	}
	return prs, nil
}
