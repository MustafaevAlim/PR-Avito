package team

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"PR/internal/client/db"
	"PR/internal/mocks"
	"PR/internal/model"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name          string
		input         *model.Team
		setupMocks    func(*mocks.MockTeamRepository, *mocks.MockTxManager)
		expectedError error
	}{
		{
			name: "успешное создание команды с участниками",
			input: &model.Team{
				TeamName: "backend-team",
				Members: []*model.TeamMember{
					{
						ID:       uuid.New(),
						Username: "john_doe",
						IsActive: true,
					},
					{
						ID:       uuid.New(),
						Username: "jane_smith",
						IsActive: true,
					},
				},
			},
			setupMocks: func(repo *mocks.MockTeamRepository, txMgr *mocks.MockTxManager) {
				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(nil)

				repo.On("CreateTeam", mock.Anything, "backend-team").Return(nil)
				repo.On("CreateMembers", mock.Anything, mock.AnythingOfType("*model.Team")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "команда с таким именем уже существует",
			input: &model.Team{
				TeamName: "existing-team",
				Members: []*model.TeamMember{
					{
						ID:       uuid.New(),
						Username: "user1",
						IsActive: true,
					},
				},
			},
			setupMocks: func(repo *mocks.MockTeamRepository, txMgr *mocks.MockTxManager) {
				pgErr := &pgconn.PgError{Code: "23505"}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrTeamExist)

				repo.On("CreateTeam", mock.Anything, "existing-team").Return(pgErr)
			},
			expectedError: ErrTeamExist,
		},
		{
			name: "ошибка при создании участников",
			input: &model.Team{
				TeamName: "new-team",
				Members: []*model.TeamMember{
					{
						ID:       uuid.New(),
						Username: "user1",
						IsActive: true,
					},
				},
			},
			setupMocks: func(repo *mocks.MockTeamRepository, txMgr *mocks.MockTxManager) {
				dbError := errors.New("foreign key constraint violation")

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(dbError)

				repo.On("CreateTeam", mock.Anything, "new-team").Return(nil)
				repo.On("CreateMembers", mock.Anything, mock.AnythingOfType("*model.Team")).Return(dbError)
			},
			expectedError: errors.New("foreign key constraint violation"),
		},
		{
			name: "команда без участников",
			input: &model.Team{
				TeamName: "empty-team",
				Members:  []*model.TeamMember{},
			},
			setupMocks: func(repo *mocks.MockTeamRepository, txMgr *mocks.MockTxManager) {
				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(nil)

				repo.On("CreateTeam", mock.Anything, "empty-team").Return(nil)
				repo.On("CreateMembers", mock.Anything, mock.AnythingOfType("*model.Team")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "общая ошибка базы данных",
			input: &model.Team{
				TeamName: "test-team",
				Members:  []*model.TeamMember{},
			},
			setupMocks: func(repo *mocks.MockTeamRepository, txMgr *mocks.MockTxManager) {
				dbError := errors.New("connection lost")

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(dbError)

				repo.On("CreateTeam", mock.Anything, "test-team").Return(dbError)
			},
			expectedError: errors.New("connection lost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockTeamRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(repo, txMgr)

			svc := NewService(repo, txMgr)

			err := svc.Create(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrTeamExist) {
					assert.ErrorIs(t, err, ErrTeamExist)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTeamByName(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		setupMocks    func(*mocks.MockTeamRepository)
		expectedTeam  *model.Team
		expectedError error
	}{
		{
			name:     "успешное получение команды",
			teamName: "backend-team",
			setupMocks: func(repo *mocks.MockTeamRepository) {
				team := &model.Team{
					TeamName: "backend-team",
					Members: []*model.TeamMember{
						{
							ID:       uuid.New(),
							Username: "john_doe",
							IsActive: true,
						},
						{
							ID:       uuid.New(),
							Username: "jane_smith",
							IsActive: false,
						},
					},
				}
				repo.On("GetTeamByName", mock.Anything, "backend-team").Return(team, nil)
			},
			expectedError: nil,
		},
		{
			name:     "команда не найдена",
			teamName: "non-existent-team",
			setupMocks: func(repo *mocks.MockTeamRepository) {
				repo.On("GetTeamByName", mock.Anything, "non-existent-team").Return(nil, pgx.ErrNoRows)
			},
			expectedTeam:  nil,
			expectedError: ErrNotFound,
		},
		{
			name:     "ошибка базы данных",
			teamName: "test-team",
			setupMocks: func(repo *mocks.MockTeamRepository) {
				repo.On("GetTeamByName", mock.Anything, "test-team").Return(nil, errors.New("database timeout"))
			},
			expectedTeam:  nil,
			expectedError: errors.New("database timeout"),
		},
		{
			name:     "команда с пустым списком участников",
			teamName: "empty-team",
			setupMocks: func(repo *mocks.MockTeamRepository) {
				team := &model.Team{
					TeamName: "empty-team",
					Members:  []*model.TeamMember{},
				}
				repo.On("GetTeamByName", mock.Anything, "empty-team").Return(team, nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockTeamRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(repo)

			svc := NewService(repo, txMgr)

			result, err := svc.GetTeamByName(context.Background(), tt.teamName)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrNotFound) {
					assert.ErrorIs(t, err, ErrNotFound)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.teamName, result.TeamName)
				assert.NotNil(t, result.Members)
			}
		})
	}
}
