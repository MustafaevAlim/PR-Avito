package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"PR/internal/client/db"
	serviceModel "PR/internal/model"
	"PR/internal/repository"
	"PR/internal/repository/user/converter"
	repoModel "PR/internal/repository/user/model"
)

type repo struct {
	db db.Client
}

func NewRepository(db db.Client) repository.UserRepository {
	return &repo{db: db}
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*serviceModel.User, error) {
	query := "SELECT * FROM users WHERE id = $1"
	var u repoModel.User
	err := r.db.DB().ScanOneContext(ctx, &u, db.Query{QueryRaw: query}, id)
	if err != nil {
		return nil, err
	}

	return converter.FromRepo(&u), nil
}

func (r *repo) GetActiveByTeam(ctx context.Context, teamName string) ([]*serviceModel.User, error) {
	var teamMates []*repoModel.User
	query := "SELECT * FROM users WHERE team_name = $1 AND is_active = true"
	err := r.db.DB().ScanAllContext(ctx, &teamMates, db.Query{QueryRaw: query}, teamName)
	if err != nil {
		return nil, err
	}
	return converter.FromRepoList(teamMates), nil
}

func (r *repo) SetActive(ctx context.Context, req *serviceModel.UserSetActive) error {
	query := `
		UPDATE users
		SET is_active = $1
		WHERE id = $2
	`
	res, err := r.db.DB().ExecContext(ctx, db.Query{QueryRaw: query}, req.IsActive, req.UserID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
