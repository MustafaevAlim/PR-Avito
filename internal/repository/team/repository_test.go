package team

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"PR/internal/client/db"
	"PR/internal/model"
	testingpkg "PR/internal/repository/testing"
)

type TeamRepositoryTestSuite struct {
	suite.Suite
	db   *testingpkg.TestDatabase
	repo *repo
}

func TestTeamRepositorySuite(t *testing.T) {
	suite.Run(t, new(TeamRepositoryTestSuite))
}

func (s *TeamRepositoryTestSuite) SetupSuite() {
	s.db = testingpkg.SetupTestDatabase(s.T())
	s.repo = &repo{db: s.db.Client}
}

func (s *TeamRepositoryTestSuite) TearDownSuite() {
	if s.db.Client != nil {
		s.db.Client.Close()
	}
}

func (s *TeamRepositoryTestSuite) TestCreateTeam_Success() {
	ctx := context.Background()
	teamName := "platform-team"

	err := s.repo.CreateTeam(ctx, teamName)
	require.NoError(s.T(), err)

	var count int
	row := s.db.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT COUNT(*) FROM teams WHERE team_name = $1",
	}, teamName)
	err = row.Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
}

func (s *TeamRepositoryTestSuite) TestCreateMembers_Success() {
	ctx := context.Background()

	teamName := "data-team"
	err := s.repo.CreateTeam(ctx, teamName)
	require.NoError(s.T(), err)

	team := &model.Team{
		TeamName: teamName,
		Members: []*model.TeamMember{
			{
				ID:       uuid.New(),
				Username: "user1",
				IsActive: true,
			},
			{
				ID:       uuid.New(),
				Username: "user2",
				IsActive: true,
			},
			{
				ID:       uuid.New(),
				Username: "user3",
				IsActive: false,
			},
		},
	}

	err = s.repo.CreateMembers(ctx, team)
	require.NoError(s.T(), err)

	var count int
	row := s.db.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT COUNT(*) FROM users WHERE team_name = $1",
	}, teamName)
	err = row.Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 3, count)
}

func (s *TeamRepositoryTestSuite) TestCreateMembers_Upsert() {
	ctx := context.Background()

	teamName := "mobile-team"
	err := s.repo.CreateTeam(ctx, teamName)
	require.NoError(s.T(), err)

	userID := uuid.New()

	team1 := &model.Team{
		TeamName: teamName,
		Members: []*model.TeamMember{
			{
				ID:       userID,
				Username: "mobile-dev",
				IsActive: true,
			},
		},
	}

	err = s.repo.CreateMembers(ctx, team1)
	require.NoError(s.T(), err)

	team2 := &model.Team{
		TeamName: teamName,
		Members: []*model.TeamMember{
			{
				ID:       userID,
				Username: "mobile-dev-updated",
				IsActive: false,
			},
		},
	}

	err = s.repo.CreateMembers(ctx, team2)
	require.NoError(s.T(), err)

	var username string
	var isActive bool
	row := s.db.Client.DB().QueryRowContext(ctx, db.Query{
		QueryRaw: "SELECT username, is_active FROM users WHERE id = $1",
	}, userID)
	err = row.Scan(&username, &isActive)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "mobile-dev-updated", username)
	assert.False(s.T(), isActive)
}

func (s *TeamRepositoryTestSuite) TestGetTeamIDByName_Success() {
	ctx := context.Background()

	teamName := "devops-team"
	err := s.repo.CreateTeam(ctx, teamName)
	require.NoError(s.T(), err)

	teamID, err := s.repo.GetTeamIDByName(ctx, teamName)
	require.NoError(s.T(), err)

	assert.NotEqual(s.T(), uuid.Nil, teamID)
}

func (s *TeamRepositoryTestSuite) TestGetTeamByName_Success() {
	ctx := context.Background()

	teamName := "qa-team"
	err := s.repo.CreateTeam(ctx, teamName)
	require.NoError(s.T(), err)

	team := &model.Team{
		TeamName: teamName,
		Members: []*model.TeamMember{
			{ID: uuid.New(), Username: "qa1", IsActive: true},
			{ID: uuid.New(), Username: "qa2", IsActive: true},
			{ID: uuid.New(), Username: "qa3", IsActive: false},
		},
	}

	err = s.repo.CreateMembers(ctx, team)
	require.NoError(s.T(), err)

	result, err := s.repo.GetTeamByName(ctx, teamName)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), teamName, result.TeamName)
	assert.Len(s.T(), result.Members, 3)

	usernames := []string{
		result.Members[0].Username,
		result.Members[1].Username,
		result.Members[2].Username,
	}
	assert.Contains(s.T(), usernames, "qa1")
	assert.Contains(s.T(), usernames, "qa2")
	assert.Contains(s.T(), usernames, "qa3")
}
