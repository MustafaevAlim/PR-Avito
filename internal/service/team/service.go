package team

import (
	"PR/internal/client/db"
	"PR/internal/repository"
	"PR/internal/service"
)

const op = "service.TeamService"

type serv struct {
	repo      repository.TeamRepository
	txManager db.TxManager
}

func NewService(
	repo repository.TeamRepository,
	txManager db.TxManager,
) service.TeamService {
	return &serv{
		repo:      repo,
		txManager: txManager,
	}
}
