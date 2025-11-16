package pr

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
		input         *model.PullRequestShort
		setupMocks    func(*mocks.MockPullRequestRepository, *mocks.MockUserRepository, *mocks.MockTxManager)
		expectedError error
	}{
		{
			name: "успешное создание PR с двумя ревьюерами",
			input: &model.PullRequestShort{
				ID:       uuid.New(),
				Name:     "feature-branch",
				AuthorID: uuid.New(),
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				authorID := uuid.New()
				teamName := "backend-team"

				author := &model.User{
					ID:       authorID,
					IsActive: true,
					TeamName: teamName,
				}

				reviewer1 := &model.User{ID: uuid.New(), IsActive: true, TeamName: teamName}
				reviewer2 := &model.User{ID: uuid.New(), IsActive: true, TeamName: teamName}

				teamMembers := []*model.User{author, reviewer1, reviewer2}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(nil)

				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(author, nil)
				userRepo.On("GetActiveByTeam", mock.Anything, teamName).Return(teamMembers, nil)
				prRepo.On("CreatePR", mock.Anything, mock.AnythingOfType("*model.PullRequest")).Return(nil)
				prRepo.On("CreatePRReviewers", mock.Anything, mock.AnythingOfType("*model.PullRequest")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "автор не найден",
			input: &model.PullRequestShort{
				ID:       uuid.New(),
				Name:     "test-pr",
				AuthorID: uuid.New(),
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrNotFound)

				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
			},
			expectedError: ErrNotFound,
		},
		{
			name: "автор не активен",
			input: &model.PullRequestShort{
				ID:       uuid.New(),
				Name:     "test-pr",
				AuthorID: uuid.New(),
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				author := &model.User{
					ID:       uuid.New(),
					IsActive: false,
					TeamName: "team",
				}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrNotActive)

				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(author, nil)
			},
			expectedError: ErrNotActive,
		},
		{
			name: "PR уже существует (duplicate key)",
			input: &model.PullRequestShort{
				ID:       uuid.New(),
				Name:     "existing-pr",
				AuthorID: uuid.New(),
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				authorID := uuid.New()
				author := &model.User{
					ID:       authorID,
					IsActive: true,
					TeamName: "team",
				}
				teamMembers := []*model.User{author, {ID: uuid.New(), IsActive: true}}

				pgErr := &pgconn.PgError{Code: "23505"}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrPRExists)

				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(author, nil)
				userRepo.On("GetActiveByTeam", mock.Anything, "team").Return(teamMembers, nil)
				prRepo.On("CreatePR", mock.Anything, mock.AnythingOfType("*model.PullRequest")).Return(pgErr)
			},
			expectedError: ErrPRExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := mocks.NewMockPullRequestRepository(t)
			userRepo := mocks.NewMockUserRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(prRepo, userRepo, txMgr)

			svc := NewService(prRepo, userRepo, txMgr)

			result, err := svc.Create(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, "OPEN", result.Status)
				assert.Len(t, result.AssignedReviewers, 2)
			}
		})
	}
}

func TestGetByReviewer(t *testing.T) {
	tests := []struct {
		name          string
		reviewerID    uuid.UUID
		setupMocks    func(*mocks.MockPullRequestRepository)
		expectedCount int
		expectedError error
	}{
		{
			name:       "успешное получение PR для ревьюера",
			reviewerID: uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prs := []*model.PullRequestShort{
					{ID: uuid.New(), Name: "pr-1", AuthorID: uuid.New()},
					{ID: uuid.New(), Name: "pr-2", AuthorID: uuid.New()},
				}
				prRepo.On("GetByReviewer", mock.Anything, mock.Anything).Return(prs, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:       "пустой список PR",
			reviewerID: uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prRepo.On("GetByReviewer", mock.Anything, mock.Anything).Return([]*model.PullRequestShort{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name:       "ошибка репозитория",
			reviewerID: uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prRepo.On("GetByReviewer", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := mocks.NewMockPullRequestRepository(t)
			userRepo := mocks.NewMockUserRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(prRepo)

			svc := NewService(prRepo, userRepo, txMgr)

			result, err := svc.GetByReviewer(context.Background(), tt.reviewerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name          string
		prID          uuid.UUID
		setupMocks    func(*mocks.MockPullRequestRepository)
		expectedError error
	}{
		{
			name: "успешный merge PR",
			prID: uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				mergedPR := &model.PullRequest{
					ID:     uuid.New(),
					Name:   "merged-pr",
					Status: "MERGED",
				}
				prRepo.On("Merge", mock.Anything, mock.Anything).Return(mergedPR, nil)
			},
			expectedError: nil,
		},
		{
			name: "PR не найден",
			prID: uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prRepo.On("Merge", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
			},
			expectedError: ErrNotFound,
		},
		{
			name: "ошибка базы данных",
			prID: uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prRepo.On("Merge", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := mocks.NewMockPullRequestRepository(t)
			userRepo := mocks.NewMockUserRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(prRepo)

			svc := NewService(prRepo, userRepo, txMgr)

			result, err := svc.Merge(context.Background(), tt.prID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrNotFound) {
					assert.ErrorIs(t, err, ErrNotFound)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "MERGED", result.Status)
			}
		})
	}
}

func TestReassignReviewers(t *testing.T) {
	tests := []struct {
		name          string
		oldID         uuid.UUID
		prID          uuid.UUID
		setupMocks    func(*mocks.MockPullRequestRepository, *mocks.MockUserRepository, *mocks.MockTxManager)
		expectedError error
	}{
		{
			name:  "успешная переназначение ревьюера",
			oldID: uuid.New(),
			prID:  uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				authorID := uuid.New()
				oldReviewerID := uuid.New()
				currentReviewerID := uuid.New()
				newReviewerID := uuid.New()

				pr := &model.PullRequest{
					ID:                uuid.New(),
					Status:            "OPEN",
					AuthorID:          authorID,
					AssignedReviewers: []uuid.UUID{oldReviewerID, currentReviewerID},
				}

				user := &model.User{
					ID:       oldReviewerID,
					TeamName: "team",
				}

				teamMembers := []*model.User{
					{ID: authorID, IsActive: true},
					{ID: oldReviewerID, IsActive: true},
					{ID: currentReviewerID, IsActive: true},
					{ID: newReviewerID, IsActive: true},
				}

				updatedPR := &model.PullRequest{
					ID:                pr.ID,
					Status:            "OPEN",
					AuthorID:          authorID,
					AssignedReviewers: []uuid.UUID{currentReviewerID, newReviewerID},
				}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(nil)

				prRepo.On("GetByID", mock.Anything, mock.Anything).Return(pr, nil).Once()
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				userRepo.On("GetActiveByTeam", mock.Anything, "team").Return(teamMembers, nil)
				prRepo.On("ReassignReviewers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				prRepo.On("GetByID", mock.Anything, mock.Anything).Return(updatedPR, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:  "PR не найден",
			oldID: uuid.New(),
			prID:  uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrNotFound)

				prRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
			},
			expectedError: ErrNotFound,
		},
		{
			name:  "PR уже смержен",
			oldID: uuid.New(),
			prID:  uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				pr := &model.PullRequest{
					ID:     uuid.New(),
					Status: "MERGED",
				}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrPRMerged)

				prRepo.On("GetByID", mock.Anything, mock.Anything).Return(pr, nil)
			},
			expectedError: ErrPRMerged,
		},
		{
			name:  "нет доступных кандидатов для переназначения",
			oldID: uuid.New(),
			prID:  uuid.New(),
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				authorID := uuid.New()
				reviewer1 := uuid.New()
				reviewer2 := uuid.New()

				pr := &model.PullRequest{
					ID:                uuid.New(),
					Status:            "OPEN",
					AuthorID:          authorID,
					AssignedReviewers: []uuid.UUID{reviewer1, reviewer2},
				}

				user := &model.User{
					ID:       reviewer1,
					TeamName: "team",
				}

				teamMembers := []*model.User{
					{ID: authorID, IsActive: true},
					{ID: reviewer1, IsActive: true},
					{ID: reviewer2, IsActive: true},
				}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrNoCandidate)

				prRepo.On("GetByID", mock.Anything, mock.Anything).Return(pr, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				userRepo.On("GetActiveByTeam", mock.Anything, "team").Return(teamMembers, nil)
			},
			expectedError: ErrNoCandidate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := mocks.NewMockPullRequestRepository(t)
			userRepo := mocks.NewMockUserRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(prRepo, userRepo, txMgr)

			svc := NewService(prRepo, userRepo, txMgr)

			result, replaceBy, err := svc.ReassignReviewers(context.Background(), tt.oldID, tt.prID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
				assert.Equal(t, uuid.UUID{}, replaceBy)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEqual(t, uuid.UUID{}, replaceBy)
			}
		})
	}
}

func TestSelectReviewers(t *testing.T) {
	authorID := uuid.New()
	member1 := &model.User{ID: uuid.New()}
	member2 := &model.User{ID: uuid.New()}
	member3 := &model.User{ID: uuid.New()}

	tests := []struct {
		name          string
		members       []*model.User
		authorID      uuid.UUID
		maxCount      int
		expectedCount int
	}{
		{
			name:          "выбор 2 ревьюеров из 3 участников",
			members:       []*model.User{{ID: authorID}, member1, member2, member3},
			authorID:      authorID,
			maxCount:      2,
			expectedCount: 2,
		},
		{
			name:          "недостаточно участников",
			members:       []*model.User{{ID: authorID}, member1},
			authorID:      authorID,
			maxCount:      2,
			expectedCount: 1,
		},
		{
			name:          "только автор в команде",
			members:       []*model.User{{ID: authorID}},
			authorID:      authorID,
			maxCount:      2,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewers := selectReviewers(tt.members, tt.authorID, tt.maxCount)
			assert.Equal(t, tt.expectedCount, len(reviewers))

			for _, reviewerID := range reviewers {
				assert.NotEqual(t, tt.authorID, reviewerID)
			}
		})
	}
}
