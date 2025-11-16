package testing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"PR/internal/client/db"
	"PR/internal/client/db/pg"
	"PR/internal/client/db/transaction"
)

const (
	dbName   = "test_db"
	user     = "test_user"
	password = "test_password"
)

type TestDatabase struct {
	Container *postgres.PostgresContainer
	Client    db.Client
	TxManager db.TxManager
	ConnStr   string
}

func SetupTestDatabase(t *testing.T) *TestDatabase {
	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	err = runMigrationsFromFiles(connStr, "../../../migrations")
	require.NoError(t, err)

	client, err := pg.New(ctx, connStr)
	require.NoError(t, err)

	txManager := transaction.NewTransactionManager(client.DB())

	return &TestDatabase{
		Container: pgContainer,
		Client:    client,
		TxManager: txManager,
		ConnStr:   connStr,
	}
}

func runMigrationsFromFiles(dbURI, migrationsPath string) error {
	migrationURL := strings.Replace(dbURI, "postgres://", "pgx5://", 1)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		migrationURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		if errSource, errDatabase := m.Close(); errSource != nil {
			log.Error().Msgf("Migrate source error: %v", errSource)
		} else if errDatabase != nil {
			log.Error().Msgf("Migrate database error: %v", errDatabase)
		}
	}()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func (td *TestDatabase) CleanupTables(t *testing.T) {
	ctx := context.Background()
	queries := []string{
		"TRUNCATE TABLE pr_reviewers CASCADE",
		"TRUNCATE TABLE prs CASCADE",
		"TRUNCATE TABLE users CASCADE",
		"TRUNCATE TABLE teams CASCADE",
	}

	for _, q := range queries {
		_, err := td.Client.DB().ExecContext(ctx, db.Query{QueryRaw: q})
		require.NoError(t, err)
	}
}
