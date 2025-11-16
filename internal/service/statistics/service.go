package statistics

import (
	"PR/internal/client/db"
	"PR/internal/repository"
	"PR/internal/service"
)

const op = "service.StatisticsService"

type serv struct {
	repo      repository.StatisticsRepository
	txManager db.TxManager
}

func NewService(
	repo repository.StatisticsRepository,
	txManager db.TxManager,
) service.StatisticsService {
	return &serv{
		repo:      repo,
		txManager: txManager,
	}
}
