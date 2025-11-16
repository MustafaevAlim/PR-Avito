package team

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"PR/internal/client/db"
	serviceModel "PR/internal/model"
	"PR/internal/repository"
	"PR/internal/repository/team/converter"
	repoModel "PR/internal/repository/team/model"
)

type repo struct {
	db db.Client
}

func NewRepository(db db.Client) repository.TeamRepository {
	return &repo{db: db}
}

func (r *repo) CreateTeam(ctx context.Context, teamName string) error {
	query := `INSERT INTO teams(id, team_name)
				VALUES ($1, $2)`
	args := []any{uuid.New(), teamName}
	_, err := r.db.DB().ExecContext(ctx, db.Query{QueryRaw: query}, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) CreateMembers(ctx context.Context, t *serviceModel.Team) error {
	batch := &pgx.Batch{}
	query := `
        INSERT INTO users(id, username, team_name, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active
    `

	for _, u := range t.Members {
		batch.Queue(query, u.ID, u.Username, t.TeamName, u.IsActive)
	}

	results := r.db.DB().SendBatch(ctx, batch)
	defer func() {
		if err := results.Close(); err != nil {
			log.Error().Msgf("Close row error: %v", err)
		}
	}()
	for i := 0; i < len(t.Members); i++ {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}

	return results.Close()
}

func (r *repo) GetTeamIDByName(ctx context.Context, name string) (uuid.UUID, error) {
	query := `SELECT id FROM teams WHERE team_name = $1`

	res := r.db.DB().QueryRowContext(ctx, db.Query{QueryRaw: query}, name)
	var id uuid.UUID
	err := res.Scan(&id)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func (r *repo) GetTeamByName(ctx context.Context, name string) (*serviceModel.Team, error) {
	var members []*repoModel.TeamMember

	query := `SELECT id, username, is_active
				FROM users
				WHERE team_name = $1`
	err := r.db.DB().ScanAllContext(ctx, &members, db.Query{QueryRaw: query}, name)
	if err != nil {
		return nil, err
	}

	return &serviceModel.Team{
		TeamName: name,
		Members:  converter.FromRepo(members),
	}, nil

}
