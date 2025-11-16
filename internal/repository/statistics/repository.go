package statistics

import (
	"context"

	"PR/internal/client/db"
	"PR/internal/model"
	"PR/internal/repository"
)

type repo struct {
	db db.Client
}

func NewRepository(db db.Client) repository.StatisticsRepository {
	return &repo{db: db}
}

func (r *repo) GetReviewerStatistics(ctx context.Context) ([]*model.ReviewerStats, error) {
	query := `
        SELECT 
            u.id as reviewer_id,
            u.username as reviewer_name,
            COUNT(pr.pr_id) as assigned_count
        FROM users u
        LEFT JOIN pr_reviewers pr ON pr.reviewer_id = u.id
        GROUP BY u.id, u.username
        ORDER BY assigned_count DESC
    `

	var stats []*model.ReviewerStats
	err := r.db.DB().ScanAllContext(ctx, &stats, db.Query{QueryRaw: query})
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *repo) GetPRStatistics(ctx context.Context) ([]*model.PRStats, error) {
	query := `
        SELECT 
            p.id as pr_id,
            p.name as pr_name,
            p.status,
            COUNT(pr.reviewer_id) as reviewer_count
        FROM prs p
        LEFT JOIN pr_reviewers pr ON pr.pr_id = p.id
        GROUP BY p.id, p.name, p.status
        ORDER BY reviewer_count DESC
    `

	var stats []*model.PRStats
	err := r.db.DB().ScanAllContext(ctx, &stats, db.Query{QueryRaw: query})
	if err != nil {
		return nil, err
	}

	return stats, nil
}
