package team

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"PR/internal/model"
)

func (s *serv) GetTeamByName(ctx context.Context, name string) (*model.Team, error) {

	t, err := s.repo.GetTeamByName(ctx, name)
	if err != nil {
		log.Error().Msgf("%s.GetTeamByName error: %v", op, err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}
