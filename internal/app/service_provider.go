package app

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"PR/internal/client/db"
	"PR/internal/client/db/pg"
	"PR/internal/client/db/transaction"
	"PR/internal/closer"
	"PR/internal/config"

	prHandler "PR/internal/api/handlers/pr"
	statHandler "PR/internal/api/handlers/statistics"
	teamHandler "PR/internal/api/handlers/team"
	userHandler "PR/internal/api/handlers/user"

	"PR/internal/repository"
	prRepo "PR/internal/repository/pr"
	statRepo "PR/internal/repository/statistics"
	teamRepo "PR/internal/repository/team"
	userRepo "PR/internal/repository/user"

	"PR/internal/service"
	prService "PR/internal/service/pr"
	statService "PR/internal/service/statistics"
	teamService "PR/internal/service/team"
	userService "PR/internal/service/user"
)

type serviceProvider struct {
	config *config.Config

	dbClient  db.Client
	txManager db.TxManager

	handlerContainer *HandlerContainer
	serviceContraier *ServiceContraier
	repoContainer    *RepoContainer
}

type HandlerContainer struct {
	User        *userHandler.UserHandler
	Team        *teamHandler.TeamHandler
	PullRequest *prHandler.PullRequestHandler
	Statistics  *statHandler.StatisticsHandler
}

type ServiceContraier struct {
	User        service.UserService
	Team        service.TeamService
	PullRequest service.PullRequestService
	Statistics  service.StatisticsService
}

type RepoContainer struct {
	User        repository.UserRepository
	Team        repository.TeamRepository
	PullRequest repository.PullRequestRepository
	Statistics  repository.StatisticsRepository
}

func (s *serviceProvider) Config() *config.Config {
	if s.config == nil {
		c, err := config.NewConfig()
		if err != nil {
			log.Fatal().Msgf("Load Config error: %v", err)
		}

		s.config = c
	}
	return s.config
}

func NewServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (s *serviceProvider) DBClient(ctx context.Context) db.Client {
	if s.dbClient == nil {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s database=%s sslmode=disable",
			s.Config().Postgre.Host, s.Config().Postgre.User, s.Config().Postgre.Password, s.Config().Postgre.DBName,
		)
		db, err := pg.New(ctx, dsn)
		if err != nil {
			log.Fatal().Msgf("Connect database error: %v", err)
		}
		err = db.DB().Ping(ctx)
		if err != nil {
			log.Fatal().Msgf("Ping database error: %v", err)
		}

		closer.Add(func() error {
			err := db.Close()
			return err
		})
		s.dbClient = db
	}

	return s.dbClient

}

func (s *serviceProvider) TxManager(ctx context.Context) db.TxManager {
	if s.txManager == nil {
		tx := transaction.NewTransactionManager(s.dbClient.DB())
		s.txManager = tx
	}
	return s.txManager
}

func (s *serviceProvider) GetRepoContainer(ctx context.Context) *RepoContainer {
	if s.repoContainer == nil {
		user := userRepo.NewRepository(s.DBClient(ctx))
		team := teamRepo.NewRepository(s.DBClient(ctx))
		pr := prRepo.NewRepository(s.DBClient(ctx))
		stat := statRepo.NewRepository(s.DBClient(ctx))

		s.repoContainer = &RepoContainer{
			User:        user,
			Team:        team,
			PullRequest: pr,
			Statistics:  stat,
		}

	}
	return s.repoContainer
}

func (s *serviceProvider) GetServiceContainer(ctx context.Context) *ServiceContraier {
	if s.serviceContraier == nil {
		user := userService.NewService(s.GetRepoContainer(ctx).User, s.TxManager(ctx))
		team := teamService.NewService(s.GetRepoContainer(ctx).Team, s.TxManager(ctx))
		pr := prService.NewService(
			s.GetRepoContainer(ctx).PullRequest,
			s.GetRepoContainer(ctx).User,
			s.TxManager(ctx),
		)
		stat := statService.NewService(s.GetRepoContainer(ctx).Statistics, s.TxManager(ctx))

		s.serviceContraier = &ServiceContraier{
			User:        user,
			Team:        team,
			PullRequest: pr,
			Statistics:  stat,
		}
	}
	return s.serviceContraier
}

func (s *serviceProvider) GetHandlerContainer(ctx context.Context) *HandlerContainer {
	if s.handlerContainer == nil {
		user := userHandler.NewUserHandler(s.GetServiceContainer(ctx).User)
		team := teamHandler.NewTeamHandler(s.GetServiceContainer(ctx).Team)
		pr := prHandler.NewPullRequestHandler(s.GetServiceContainer(ctx).PullRequest)
		stat := statHandler.NewHandler(s.GetServiceContainer(ctx).Statistics)

		s.handlerContainer = &HandlerContainer{
			User:        user,
			Team:        team,
			PullRequest: pr,
			Statistics:  stat,
		}
	}
	return s.handlerContainer
}
