package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"PR/internal/closer"
)

type App struct {
	serviceProvider *serviceProvider
	server          *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}
	return a, nil

}
func (a *App) Run() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	return a.runServer()

}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initServiceProvider,
		a.initServer,
		a.initLogger,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.serviceProvider = NewServiceProvider()
	return nil
}

func (a *App) initServer(ctx context.Context) error {
	handler := a.serviceProvider.GetHandlerContainer(ctx)
	engine := gin.New()
	setupRoutes(handler, engine)

	server := &http.Server{
		Addr:              ":" + a.serviceProvider.Config().Server.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}
	a.server = server

	closer.Add(server.Close)

	return nil
}

func (a *App) runServer() error {
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func setupRoutes(h *HandlerContainer, e *gin.Engine) {
	e.Use(gin.Logger(), gin.Recovery())

	e.StaticFS("/swagger", http.Dir("./api/dist"))

	// Отдача swagger.yaml
	e.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("./api/openapi.yaml")
	})

	e.POST("/team/add", h.Team.Create)
	e.GET("/team/get", h.Team.GetTeamByName)

	e.POST("/pullRequest/create", h.PullRequest.Create)
	e.POST("/pullRequest/merge", h.PullRequest.Merge)
	e.POST("/pullRequest/reassign", h.PullRequest.Reassign)

	e.POST("/users/setIsActive", h.User.SetActive)
	e.GET("/users/getReview", h.PullRequest.GetByReviewer)

	stats := e.Group("/statistics")
	{
		stats.GET("/reviewers", h.Statistics.GetReviewerStats)
		stats.GET("/prs", h.Statistics.GetPRStats)
	}

}
