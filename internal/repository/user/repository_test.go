package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"PR/internal/client/db"
	"PR/internal/model"
	testingpkg "PR/internal/repository/testing"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *testingpkg.TestDatabase
	repo *repo
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	s.db = testingpkg.SetupTestDatabase(s.T())
	s.repo = &repo{db: s.db.Client}
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	if s.db.Client != nil {
		s.db.Client.Close()
	}
}
func (s *UserRepositoryTestSuite) SetupTest() {
	s.db.CleanupTables(s.T())
	s.seedTestData()
}

func (s *UserRepositoryTestSuite) seedTestData() {
	ctx := context.Background()

	teamID := uuid.New()
	_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO teams(id, team_name) VALUES ($1, $2)",
	}, teamID, "test-team")
	require.NoError(s.T(), err)
}

func (s *UserRepositoryTestSuite) TestGetByID_Success() {
	ctx := context.Background()

	userID := uuid.New()
	username := "test-user"

	_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO users(id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
	}, userID, username, "test-team", true)
	require.NoError(s.T(), err)

	user, err := s.repo.GetByID(ctx, userID)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), userID, user.ID)
	assert.Equal(s.T(), username, user.Username)
	assert.True(s.T(), user.IsActive)
}

func (s *UserRepositoryTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()

	nonExistentID := uuid.New()
	_, err := s.repo.GetByID(ctx, nonExistentID)

	assert.Error(s.T(), err)
}

func (s *UserRepositoryTestSuite) TestGetActiveByTeam_Success() {
	ctx := context.Background()

	users := []struct {
		username string
		isActive bool
	}{
		{"active1", true},
		{"active2", true},
		{"inactive1", false},
		{"inactive2", false},
	}

	for _, u := range users {
		_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
			QueryRaw: "INSERT INTO users(id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
		}, uuid.New(), u.username, "test-team", u.isActive)
		require.NoError(s.T(), err)
	}

	activeUsers, err := s.repo.GetActiveByTeam(ctx, "test-team")
	require.NoError(s.T(), err)

	assert.Len(s.T(), activeUsers, 2)

	for _, user := range activeUsers {
		assert.True(s.T(), user.IsActive)
	}
}

func (s *UserRepositoryTestSuite) TestGetActiveByTeam_NoActiveUsers() {
	ctx := context.Background()

	_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO users(id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
	}, uuid.New(), "inactive", "test-team", false)
	require.NoError(s.T(), err)

	activeUsers, err := s.repo.GetActiveByTeam(ctx, "test-team")
	require.NoError(s.T(), err)

	assert.Empty(s.T(), activeUsers)
}

func (s *UserRepositoryTestSuite) TestSetActive_Success() {
	ctx := context.Background()

	userID := uuid.New()
	_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO users(id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
	}, userID, "test-user", "test-team", true)
	require.NoError(s.T(), err)

	req := &model.UserSetActive{
		UserID:   userID,
		IsActive: false,
	}

	err = s.repo.SetActive(ctx, req)
	require.NoError(s.T(), err)

	var isActive bool
	row := s.db.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT is_active FROM users WHERE id = $1",
	}, userID)
	err = row.Scan(&isActive)
	require.NoError(s.T(), err)

	assert.False(s.T(), isActive)
}

func (s *UserRepositoryTestSuite) TestSetActive_UserNotFound() {
	ctx := context.Background()

	req := &model.UserSetActive{
		UserID:   uuid.New(),
		IsActive: false,
	}

	err := s.repo.SetActive(ctx, req)

	assert.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, pgx.ErrNoRows)
}

func (s *UserRepositoryTestSuite) TestSetActive_ToggleMultipleTimes() {
	ctx := context.Background()

	userID := uuid.New()
	_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO users(id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
	}, userID, "toggle-user", "test-team", true)
	require.NoError(s.T(), err)

	err = s.repo.SetActive(ctx, &model.UserSetActive{UserID: userID, IsActive: false})
	require.NoError(s.T(), err)

	err = s.repo.SetActive(ctx, &model.UserSetActive{UserID: userID, IsActive: true})
	require.NoError(s.T(), err)

	user, err := s.repo.GetByID(ctx, userID)
	require.NoError(s.T(), err)
	assert.True(s.T(), user.IsActive)
}
