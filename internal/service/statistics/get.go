package statistics

import (
	"context"

	"github.com/rs/zerolog/log"

	"PR/internal/model"
)

func (s *serv) GetReviewerStatistics(ctx context.Context) ([]*model.ReviewerStats, error) {
	list, err := s.repo.GetReviewerStatistics(ctx)
	if err != nil {
		log.Error().Msgf("%s.GetReviewerStatistics error: %v", op, err)
		return nil, err
	}
	return list, nil
}

func (s *serv) GetPRStatistics(ctx context.Context) ([]*model.PRStats, error) {
	list, err := s.repo.GetPRStatistics(ctx)
	if err != nil {
		log.Error().Msgf("%s.GetPRStatistics error: %v", op, err)
		return nil, err
	}
	return list, nil
}
