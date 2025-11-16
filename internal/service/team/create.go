package team

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"

	"PR/internal/model"
)

func (s *serv) Create(ctx context.Context, t *model.Team) error {

	err := s.txManager.ReadCommited(ctx, func(ctx context.Context) error {
		var errTx error

		errTx = s.repo.CreateTeam(ctx, t.TeamName)
		if errTx != nil {
			var pgErr *pgconn.PgError
			if errors.As(errTx, &pgErr) {
				if pgErr.Code == "23505" {
					return ErrTeamExist
				}
			}
			return errTx
		}

		errTx = s.repo.CreateMembers(ctx, t)
		if errTx != nil {
			return errTx
		}
		return nil
	})

	if err != nil {
		log.Error().Msgf("%s.Create error: %v", op, err)
		return err
	}
	return nil
}
