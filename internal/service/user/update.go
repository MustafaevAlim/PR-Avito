package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"PR/internal/model"
)

func (s *serv) SetActive(ctx context.Context, req *model.UserSetActive) (*model.User, error) {
	var u *model.User

	err := s.txManager.ReadCommited(ctx, func(ctx context.Context) error {
		var errTx error
		errTx = s.repo.SetActive(ctx, req)
		if errTx != nil {
			if errors.Is(errTx, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return errTx
		}

		u, errTx = s.repo.GetByID(ctx, req.UserID)
		if errTx != nil {
			return errTx
		}
		return nil
	})

	if err != nil {
		log.Error().Msgf("%s.SetActive error: %v", op, err)

		return nil, err
	}
	return u, nil
}
