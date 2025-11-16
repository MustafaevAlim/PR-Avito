package statistics

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"PR/internal/client/db"
	testingpkg "PR/internal/repository/testing"
)

type StatisticsRepositoryTestSuite struct {
	suite.Suite
	db   *testingpkg.TestDatabase
	repo *repo
}

func TestStatisticsRepositorySuite(t *testing.T) {
	suite.Run(t, new(StatisticsRepositoryTestSuite))
}
func (s *StatisticsRepositoryTestSuite) SetupSuite() {
	s.db = testingpkg.SetupTestDatabase(s.T())
	s.repo = &repo{db: s.db.Client}
}

func (s *StatisticsRepositoryTestSuite) TearDownSuite() {
	if s.db.Client != nil {
		s.db.Client.Close()
	}
}
func (s *StatisticsRepositoryTestSuite) SetupTest() {
	s.db.CleanupTables(s.T())
	s.seedTestData()
}

func (s *StatisticsRepositoryTestSuite) seedTestData() {
	ctx := context.Background()

	teamID := uuid.New()
	_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO teams(id, team_name) VALUES ($1, $2)",
	}, teamID, "stats-team")
	require.NoError(s.T(), err)

	s.createUser("author1", "stats-team", true)
	s.createUser("reviewer1", "stats-team", true)
	s.createUser("reviewer2", "stats-team", true)
	s.createUser("reviewer3", "stats-team", true)
}

func (s *StatisticsRepositoryTestSuite) createUser(username, teamName string, isActive bool) uuid.UUID {
	ctx := context.Background()
	userID := uuid.New()

	_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO users(id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
	}, userID, username, teamName, isActive)
	require.NoError(s.T(), err)

	return userID
}

func (s *StatisticsRepositoryTestSuite) getUserID(username string) uuid.UUID {
	ctx := context.Background()
	var userID uuid.UUID

	row := s.db.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT id FROM users WHERE username = $1",
	}, username)

	err := row.Scan(&userID)
	require.NoError(s.T(), err)

	return userID
}

func (s *StatisticsRepositoryTestSuite) TestGetReviewerStatistics_Success() {
	ctx := context.Background()

	authorID := s.getUserID("author1")
	reviewer1ID := s.getUserID("reviewer1")
	reviewer2ID := s.getUserID("reviewer2")
	reviewer3ID := s.getUserID("reviewer3")

	pr1ID := uuid.New()
	pr2ID := uuid.New()
	pr3ID := uuid.New()

	prs := []struct {
		id       uuid.UUID
		name     string
		authorID uuid.UUID
	}{
		{pr1ID, "PR 1", authorID},
		{pr2ID, "PR 2", authorID},
		{pr3ID, "PR 3", authorID},
	}

	for _, pr := range prs {
		_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
			QueryRaw: "INSERT INTO prs(id, name, author_id, status, created_at) VALUES ($1, $2, $3, $4, NOW())",
		}, pr.id, pr.name, pr.authorID, "OPEN")
		require.NoError(s.T(), err)
	}

	reviewerAssignments := []struct {
		prID       uuid.UUID
		reviewerID uuid.UUID
	}{
		{pr1ID, reviewer1ID},
		{pr2ID, reviewer1ID},
		{pr2ID, reviewer2ID},
		{pr3ID, reviewer2ID},
	}

	for _, assignment := range reviewerAssignments {
		_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
			QueryRaw: "INSERT INTO pr_reviewers(pr_id, reviewer_id, assigned_at) VALUES ($1, $2, NOW())",
		}, assignment.prID, assignment.reviewerID)
		require.NoError(s.T(), err)
	}

	stats, err := s.repo.GetReviewerStatistics(ctx)
	require.NoError(s.T(), err)

	assert.GreaterOrEqual(s.T(), len(stats), 3)

	if len(stats) >= 2 {
		assert.GreaterOrEqual(s.T(), stats[0].AssignedCount, stats[1].AssignedCount)
	}

	statsMap := make(map[uuid.UUID]int)
	for _, stat := range stats {
		statsMap[stat.ReviewerID] = stat.AssignedCount
	}

	assert.Equal(s.T(), 2, statsMap[reviewer1ID])
	assert.Equal(s.T(), 2, statsMap[reviewer2ID])
	assert.Equal(s.T(), 0, statsMap[reviewer3ID])
}

func (s *StatisticsRepositoryTestSuite) TestGetPRStatistics_Success() {
	ctx := context.Background()

	authorID := s.getUserID("author1")
	reviewer1ID := s.getUserID("reviewer1")
	reviewer2ID := s.getUserID("reviewer2")

	pr1ID := uuid.New()
	pr2ID := uuid.New()
	pr3ID := uuid.New()

	prs := []struct {
		id     uuid.UUID
		name   string
		status string
	}{
		{pr1ID, "PR with 2 reviewers", "OPEN"},
		{pr2ID, "PR with 1 reviewer", "IN_REVIEW"},
		{pr3ID, "PR with 0 reviewers", "MERGED"},
	}

	for _, pr := range prs {
		_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
			QueryRaw: "INSERT INTO prs(id, name, author_id, status, created_at) VALUES ($1, $2, $3, $4, NOW())",
		}, pr.id, pr.name, authorID, pr.status)
		require.NoError(s.T(), err)
	}

	reviewerAssignments := []struct {
		prID       uuid.UUID
		reviewerID uuid.UUID
	}{
		{pr1ID, reviewer1ID},
		{pr1ID, reviewer2ID},
		{pr2ID, reviewer1ID},
	}

	for _, assignment := range reviewerAssignments {
		_, err := s.db.Client.DB().ExecContext(ctx, db.Query{
			QueryRaw: "INSERT INTO pr_reviewers(pr_id, reviewer_id, assigned_at) VALUES ($1, $2, NOW())",
		}, assignment.prID, assignment.reviewerID)
		require.NoError(s.T(), err)
	}

	stats, err := s.repo.GetPRStatistics(ctx)
	require.NoError(s.T(), err)

	assert.Len(s.T(), stats, 3)

	assert.GreaterOrEqual(s.T(), stats[0].ReviewerCount, stats[1].ReviewerCount)
	assert.GreaterOrEqual(s.T(), stats[1].ReviewerCount, stats[2].ReviewerCount)

	statsMap := make(map[uuid.UUID]*struct {
		name          string
		status        string
		reviewerCount int
	})

	for _, stat := range stats {
		statsMap[stat.PRID] = &struct {
			name          string
			status        string
			reviewerCount int
		}{
			name:          stat.PRName,
			status:        stat.Status,
			reviewerCount: stat.ReviewerCount,
		}
	}

	assert.Equal(s.T(), 2, statsMap[pr1ID].reviewerCount)
	assert.Equal(s.T(), 1, statsMap[pr2ID].reviewerCount)
	assert.Equal(s.T(), 0, statsMap[pr3ID].reviewerCount)
}

func (s *StatisticsRepositoryTestSuite) TestGetReviewerStatistics_EmptyDatabase() {
	ctx := context.Background()

	stats, err := s.repo.GetReviewerStatistics(ctx)
	require.NoError(s.T(), err)

	assert.NotEmpty(s.T(), stats)
}

func (s *StatisticsRepositoryTestSuite) TestGetPRStatistics_EmptyDatabase() {
	ctx := context.Background()

	stats, err := s.repo.GetPRStatistics(ctx)
	require.NoError(s.T(), err)

	assert.Empty(s.T(), stats)
}
