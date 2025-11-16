package pr

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"PR/internal/client/db"
	serviceModel "PR/internal/model"
	"PR/internal/repository"
	"PR/internal/repository/pr/converter"
	repoModel "PR/internal/repository/pr/model"
)

type repo struct {
	db db.Client
}

func NewRepository(db db.Client) repository.PullRequestRepository {
	return &repo{db: db}
}

func (r *repo) CreatePR(ctx context.Context, pr *serviceModel.PullRequest) error {
	query := `INSERT INTO prs(id, name, author_id, status, created_at)
				VALUES ($1, $2, $3, $4, NOW())`

	args := []any{pr.ID, pr.Name, pr.AuthorID, pr.Status}
	_, err := r.db.DB().ExecContext(ctx, db.Query{QueryRaw: query}, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) CreatePRReviewers(ctx context.Context, pr *serviceModel.PullRequest) error {
	query := `INSERT INTO pr_reviewers (pr_id, reviewer_id, assigned_at)
				VALUES($1, $2, NOW())`

	for _, revID := range pr.AssignedReviewers {
		args := []any{pr.ID, revID}
		_, err := r.db.DB().ExecContext(ctx, db.Query{QueryRaw: query}, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*serviceModel.PullRequest, error) {
	query := `
        SELECT 
            p.id, 
            p.name, 
            p.author_id, 
            p.status, 
            p.created_at, 
            p.merged_at,
            COALESCE(
                array_agg(pr.reviewer_id) FILTER (WHERE pr.reviewer_id IS NOT NULL), 
                '{}'
            ) as reviewers
        FROM prs p
        LEFT JOIN pr_reviewers pr ON pr.pr_id = p.id
        WHERE p.id = $1
        GROUP BY p.id, p.name, p.author_id, p.status, p.created_at, p.merged_at
    `

	var pr repoModel.PullRequest
	var reviewers []uuid.UUID

	err := r.db.DB().QueryRowContext(ctx, db.Query{QueryRaw: query}, id).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
		&reviewers,
	)

	if err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return converter.FromRepo(&pr), nil
}

func (r *repo) GetViewers(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	query := `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pr_id = $1
	`
	err := r.db.DB().ScanAllContext(ctx, &userIDs, db.Query{QueryRaw: query}, id)
	if err != nil {
		return nil, err
	}
	return userIDs, nil

}

func (r *repo) Merge(ctx context.Context, id uuid.UUID) (*serviceModel.PullRequest, error) {
	checkQuery := `SELECT status FROM prs WHERE id = $1`

	var currentStatus string
	err := r.db.DB().QueryRowContext(ctx, db.Query{QueryRaw: checkQuery}, id).Scan(&currentStatus)
	if err != nil {
		return nil, err
	}

	if currentStatus == "MERGED" {
		return r.GetByID(ctx, id)
	}

	query := `
        WITH updated AS (
            UPDATE prs
            SET status = 'MERGED', merged_at = NOW()
            WHERE id = $1
            RETURNING id, name, author_id, status, created_at, merged_at
        )
        SELECT 
            u.id, u.name, u.author_id, u.status, u.created_at, u.merged_at,
            COALESCE(array_agg(pr.reviewer_id) FILTER (WHERE pr.reviewer_id IS NOT NULL), '{}')
        FROM updated u
        LEFT JOIN pr_reviewers pr ON pr.pr_id = u.id
        GROUP BY u.id, u.name, u.author_id, u.status, u.created_at, u.merged_at
    `

	var pr repoModel.PullRequest
	var reviewers []uuid.UUID

	err = r.db.DB().QueryRowContext(ctx, db.Query{QueryRaw: query}, id).Scan(
		&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status,
		&pr.CreatedAt, &pr.MergedAt, &reviewers,
	)
	if err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return converter.FromRepo(&pr), nil

}

func (r *repo) ReassignReviewers(ctx context.Context, prID, oldID, newID uuid.UUID) error {
	query := `
        UPDATE pr_reviewers
        SET reviewer_id = $1
        WHERE reviewer_id = $2 AND pr_id = $3
    `

	result, err := r.db.DB().ExecContext(ctx, db.Query{QueryRaw: query}, newID, oldID, prID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil

}

func (r *repo) GetByReviewer(ctx context.Context, userID uuid.UUID) ([]*serviceModel.PullRequestShort, error) {
	query := `
		SELECT p.id, p.name, p.author_id, p.status
		FROM pr_reviewers pr
		INNER JOIN prs p ON p.id = pr.pr_id
		WHERE pr.reviewer_id = $1
	`
	var prs []*repoModel.PullRequestShort
	err := r.db.DB().ScanAllContext(ctx, &prs, db.Query{QueryRaw: query}, userID)
	if err != nil {
		return nil, err
	}

	return converter.FromRepoShortList(prs), nil

}
