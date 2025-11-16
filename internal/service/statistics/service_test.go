package statistics

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"PR/internal/mocks"
	"PR/internal/model"
)

func TestGetReviewerStatistics(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockStatisticsRepository)
		expectedCount int
		expectedError error
	}{
		{
			name: "успешное получение статистики ревьюеров",
			setupMocks: func(repo *mocks.MockStatisticsRepository) {
				stats := []*model.ReviewerStats{
					{
						ReviewerID:    uuid.New(),
						ReviewerName:  "John Doe",
						AssignedCount: 10,
					},
					{
						ReviewerID:    uuid.New(),
						ReviewerName:  "Jane Smith",
						AssignedCount: 5,
					},
				}
				repo.On("GetReviewerStatistics", mock.Anything).Return(stats, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "пустая статистика",
			setupMocks: func(repo *mocks.MockStatisticsRepository) {
				repo.On("GetReviewerStatistics", mock.Anything).Return([]*model.ReviewerStats{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name: "ошибка репозитория",
			setupMocks: func(repo *mocks.MockStatisticsRepository) {
				repo.On("GetReviewerStatistics", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockStatisticsRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(repo)

			svc := NewService(repo, txMgr)

			result, err := svc.GetReviewerStatistics(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)

				if tt.expectedCount > 0 {
					for _, stat := range result {
						assert.NotEqual(t, uuid.Nil, stat.ReviewerID)
						assert.NotEmpty(t, stat.ReviewerName)
						assert.GreaterOrEqual(t, stat.AssignedCount, 0)
					}
				}
			}
		})
	}
}

func TestGetPRStatistics(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockStatisticsRepository)
		expectedCount int
		expectedError error
	}{
		{
			name: "успешное получение статистики PR",
			setupMocks: func(repo *mocks.MockStatisticsRepository) {
				stats := []*model.PRStats{
					{
						PRID:          uuid.New(),
						PRName:        "feature/add-auth",
						ReviewerCount: 2,
						Status:        "OPEN",
					},
					{
						PRID:          uuid.New(),
						PRName:        "bugfix/login",
						ReviewerCount: 3,
						Status:        "MERGED",
					},
					{
						PRID:          uuid.New(),
						PRName:        "feature/api",
						ReviewerCount: 1,
						Status:        "OPEN",
					},
				}
				repo.On("GetPRStatistics", mock.Anything).Return(stats, nil)
			},
			expectedCount: 3,
			expectedError: nil,
		},
		{
			name: "пустая статистика",
			setupMocks: func(repo *mocks.MockStatisticsRepository) {
				repo.On("GetPRStatistics", mock.Anything).Return([]*model.PRStats{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name: "ошибка репозитория",
			setupMocks: func(repo *mocks.MockStatisticsRepository) {
				repo.On("GetPRStatistics", mock.Anything).Return(nil, errors.New("connection timeout"))
			},
			expectedError: errors.New("connection timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockStatisticsRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(repo)

			svc := NewService(repo, txMgr)

			result, err := svc.GetPRStatistics(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)

				if tt.expectedCount > 0 {
					for _, stat := range result {
						assert.NotEqual(t, uuid.Nil, stat.PRID)
						assert.NotEmpty(t, stat.PRName)
						assert.GreaterOrEqual(t, stat.ReviewerCount, 0)
						assert.NotEmpty(t, stat.Status)
					}
				}
			}
		})
	}
}
