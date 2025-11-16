package pr

import (
	"PR/internal/client/db"
	"PR/internal/repository"
	"PR/internal/service"
)

const op = "service.PullRequestService"

type serv struct {
	pullRequestRepo repository.PullRequestRepository
	userRepo        repository.UserRepository
	txManager       db.TxManager
}

func NewService(
	pullRequestRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	txManager db.TxManager,
) service.PullRequestService {
	return &serv{
		pullRequestRepo: pullRequestRepo,
		userRepo:        userRepo,
		txManager:       txManager,
	}
}
