package user

import (
	"PR/internal/client/db"
	"PR/internal/repository"
	"PR/internal/service"
)

const op = "service.UserService"

type serv struct {
	repo      repository.UserRepository
	txManager db.TxManager
}

func NewService(
	repo repository.UserRepository,
	txManager db.TxManager,
) service.UserService {
	return &serv{
		repo:      repo,
		txManager: txManager,
	}
}
