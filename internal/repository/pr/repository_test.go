package pr

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"PR/internal/client/db"
	"PR/internal/model"
	testingpkg "PR/internal/repository/testing"
)

type PullRequestRepositoryTestSuite struct {
	suite.Suite
	testDB *testingpkg.TestDatabase
	repo   *repo
}

func TestPullRequestRepositorySuite(t *testing.T) {
	suite.Run(t, new(PullRequestRepositoryTestSuite))
}

func (s *PullRequestRepositoryTestSuite) SetupSuite() {
	s.testDB = testingpkg.SetupTestDatabase(s.T())
	s.repo = &repo{db: s.testDB.Client}
}

func (s *PullRequestRepositoryTestSuite) TearDownSuite() {
	if s.testDB.Client != nil {
		s.testDB.Client.Close()
	}
}

func (s *PullRequestRepositoryTestSuite) SetupTest() {
	s.testDB.CleanupTables(s.T())
	s.seedTestData()
}

func (s *PullRequestRepositoryTestSuite) seedTestData() {
	ctx := context.Background()

	teamID := uuid.New()
	_, err := s.testDB.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO teams(id, team_name) VALUES ($1, $2)",
	}, teamID, "backend-team")
	require.NoError(s.T(), err)

	s.createTestUser("author-1", "backend-team", true)
	s.createTestUser("reviewer-1", "backend-team", true)
	s.createTestUser("reviewer-2", "backend-team", true)
}

func (s *PullRequestRepositoryTestSuite) createTestUser(username, teamName string, isActive bool) uuid.UUID {
	ctx := context.Background()
	userID := uuid.New()

	_, err := s.testDB.Client.DB().ExecContext(ctx, db.Query{
		QueryRaw: "INSERT INTO users(id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
	}, userID, username, teamName, isActive)
	require.NoError(s.T(), err)

	return userID
}

func (s *PullRequestRepositoryTestSuite) getUserIDByUsername(username string) uuid.UUID {
	ctx := context.Background()
	var userID uuid.UUID

	row := s.testDB.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT id FROM users WHERE username = $1",
	}, username)

	err := row.Scan(&userID)
	require.NoError(s.T(), err)

	return userID
}

func (s *PullRequestRepositoryTestSuite) TestCreatePR_Success() {
	ctx := context.Background()

	authorID := s.getUserIDByUsername("author-1")
	prID := uuid.New()

	pr := &model.PullRequest{
		ID:       prID,
		Name:     "Feature: Add new endpoint",
		AuthorID: authorID,
		Status:   "OPEN",
	}

	err := s.repo.CreatePR(ctx, pr)
	require.NoError(s.T(), err)

	var count int
	row := s.testDB.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT COUNT(*) FROM prs WHERE id = $1",
	}, prID)
	err = row.Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
}

func (s *PullRequestRepositoryTestSuite) TestCreatePRReviewers_Success() {
	ctx := context.Background()

	authorID := s.getUserIDByUsername("author-1")
	reviewer1ID := s.getUserIDByUsername("reviewer-1")
	reviewer2ID := s.getUserIDByUsername("reviewer-2")

	prID := uuid.New()
	pr := &model.PullRequest{
		ID:                prID,
		Name:              "Feature: Add logging",
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []uuid.UUID{reviewer1ID, reviewer2ID},
	}

	err := s.repo.CreatePR(ctx, pr)
	require.NoError(s.T(), err)

	err = s.repo.CreatePRReviewers(ctx, pr)
	require.NoError(s.T(), err)

	var count int
	row := s.testDB.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT COUNT(*) FROM pr_reviewers WHERE pr_id = $1",
	}, prID)
	err = row.Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, count)
}

func (s *PullRequestRepositoryTestSuite) TestGetByID_Success() {
	ctx := context.Background()

	authorID := s.getUserIDByUsername("author-1")
	reviewer1ID := s.getUserIDByUsername("reviewer-1")
	reviewer2ID := s.getUserIDByUsername("reviewer-2")

	prID := uuid.New()
	pr := &model.PullRequest{
		ID:                prID,
		Name:              "Fix: Bug in handler",
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []uuid.UUID{reviewer1ID, reviewer2ID},
	}

	err := s.repo.CreatePR(ctx, pr)
	require.NoError(s.T(), err)
	err = s.repo.CreatePRReviewers(ctx, pr)
	require.NoError(s.T(), err)

	result, err := s.repo.GetByID(ctx, prID)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), prID, result.ID)
	assert.Equal(s.T(), "Fix: Bug in handler", result.Name)
	assert.Equal(s.T(), authorID, result.AuthorID)
	assert.Equal(s.T(), "OPEN", result.Status)
	assert.Len(s.T(), result.AssignedReviewers, 2)
	assert.Contains(s.T(), result.AssignedReviewers, reviewer1ID)
	assert.Contains(s.T(), result.AssignedReviewers, reviewer2ID)
}

func (s *PullRequestRepositoryTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()

	nonExistentID := uuid.New()
	_, err := s.repo.GetByID(ctx, nonExistentID)

	assert.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, pgx.ErrNoRows)
}

func (s *PullRequestRepositoryTestSuite) TestMerge_Success() {
	ctx := context.Background()

	authorID := s.getUserIDByUsername("author-1")
	reviewer1ID := s.getUserIDByUsername("reviewer-1")

	prID := uuid.New()
	pr := &model.PullRequest{
		ID:                prID,
		Name:              "Feature: Integration tests",
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []uuid.UUID{reviewer1ID},
	}

	err := s.repo.CreatePR(ctx, pr)
	require.NoError(s.T(), err)
	err = s.repo.CreatePRReviewers(ctx, pr)
	require.NoError(s.T(), err)

	merged, err := s.repo.Merge(ctx, prID)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "MERGED", merged.Status)
	assert.NotNil(s.T(), merged.MergedAt)
	assert.WithinDuration(s.T(), time.Now(), *merged.MergedAt, 5*time.Second)
}

func (s *PullRequestRepositoryTestSuite) TestMerge_AlreadyMerged() {
	ctx := context.Background()

	authorID := s.getUserIDByUsername("author-1")
	prID := uuid.New()

	pr := &model.PullRequest{
		ID:       prID,
		Name:     "Feature: Already merged",
		AuthorID: authorID,
		Status:   "OPEN",
	}

	err := s.repo.CreatePR(ctx, pr)
	require.NoError(s.T(), err)

	_, err = s.repo.Merge(ctx, prID)
	require.NoError(s.T(), err)

	merged, err := s.repo.Merge(ctx, prID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "MERGED", merged.Status)
}

func (s *PullRequestRepositoryTestSuite) TestReassignReviewers_Success() {
	ctx := context.Background()

	authorID := s.getUserIDByUsername("author-1")
	oldReviewerID := s.getUserIDByUsername("reviewer-1")
	newReviewerID := s.getUserIDByUsername("reviewer-2")

	prID := uuid.New()
	pr := &model.PullRequest{
		ID:                prID,
		Name:              "Feature: Reassignment test",
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []uuid.UUID{oldReviewerID},
	}

	err := s.repo.CreatePR(ctx, pr)
	require.NoError(s.T(), err)
	err = s.repo.CreatePRReviewers(ctx, pr)
	require.NoError(s.T(), err)

	err = s.repo.ReassignReviewers(ctx, prID, oldReviewerID, newReviewerID)
	require.NoError(s.T(), err)

	reviewers, err := s.repo.GetViewers(ctx, prID)
	require.NoError(s.T(), err)

	assert.Len(s.T(), reviewers, 1)
	assert.Equal(s.T(), newReviewerID, reviewers[0])
}

func (s *PullRequestRepositoryTestSuite) TestReassignReviewers_NoRowsAffected() {
	ctx := context.Background()

	prID := uuid.New()
	oldReviewerID := uuid.New()
	newReviewerID := uuid.New()

	err := s.repo.ReassignReviewers(ctx, prID, oldReviewerID, newReviewerID)

	assert.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, pgx.ErrNoRows)
}

func (s *PullRequestRepositoryTestSuite) TestGetByReviewer_Success() {
	ctx := context.Background()

	authorID := s.getUserIDByUsername("author-1")
	reviewerID := s.getUserIDByUsername("reviewer-1")

	pr1ID := uuid.New()
	pr1 := &model.PullRequest{
		ID:                pr1ID,
		Name:              "PR 1",
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []uuid.UUID{reviewerID},
	}

	pr2ID := uuid.New()
	pr2 := &model.PullRequest{
		ID:                pr2ID,
		Name:              "PR 2",
		AuthorID:          authorID,
		Status:            "IN_REVIEW",
		AssignedReviewers: []uuid.UUID{reviewerID},
	}

	for _, pr := range []*model.PullRequest{pr1, pr2} {
		err := s.repo.CreatePR(ctx, pr)
		require.NoError(s.T(), err)
		err = s.repo.CreatePRReviewers(ctx, pr)
		require.NoError(s.T(), err)
	}

	prs, err := s.repo.GetByReviewer(ctx, reviewerID)
	require.NoError(s.T(), err)

	assert.Len(s.T(), prs, 2)

	prIDs := []uuid.UUID{prs[0].ID, prs[1].ID}
	assert.Contains(s.T(), prIDs, pr1ID)
	assert.Contains(s.T(), prIDs, pr2ID)
}
