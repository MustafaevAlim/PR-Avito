package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"PR/internal/client/db"
	"PR/internal/mocks"
	"PR/internal/model"
)

func TestSetActive(t *testing.T) {
	tests := []struct {
		name          string
		input         *model.UserSetActive
		setupMocks    func(*mocks.MockUserRepository, *mocks.MockTxManager)
		expectedError error
		checkResult   func(*testing.T, *model.User)
	}{
		{
			name: "успешная активация пользователя",
			input: &model.UserSetActive{
				UserID:   uuid.New(),
				IsActive: true,
			},
			setupMocks: func(repo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				userID := uuid.New()
				updatedUser := &model.User{
					ID:       userID,
					Username: "john_doe",
					IsActive: true,
					TeamName: "backend-team",
				}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(nil)

				repo.On("SetActive", mock.Anything, mock.AnythingOfType("*model.UserSetActive")).Return(nil)
				repo.On("GetByID", mock.Anything, mock.Anything).Return(updatedUser, nil)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, u *model.User) {
				assert.NotNil(t, u)
				assert.True(t, u.IsActive)
			},
		},
		{
			name: "успешная деактивация пользователя",
			input: &model.UserSetActive{
				UserID:   uuid.New(),
				IsActive: false,
			},
			setupMocks: func(repo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				userID := uuid.New()
				updatedUser := &model.User{
					ID:       userID,
					Username: "jane_smith",
					IsActive: false,
					TeamName: "frontend-team",
				}

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(nil)

				repo.On("SetActive", mock.Anything, mock.AnythingOfType("*model.UserSetActive")).Return(nil)
				repo.On("GetByID", mock.Anything, mock.Anything).Return(updatedUser, nil)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, u *model.User) {
				assert.NotNil(t, u)
				assert.False(t, u.IsActive)
			},
		},
		{
			name: "пользователь не найден",
			input: &model.UserSetActive{
				UserID:   uuid.New(),
				IsActive: true,
			},
			setupMocks: func(repo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrNotFound)

				repo.On("SetActive", mock.Anything, mock.AnythingOfType("*model.UserSetActive")).Return(pgx.ErrNoRows)
			},
			expectedError: ErrNotFound,
			checkResult: func(t *testing.T, u *model.User) {
				assert.Nil(t, u)
			},
		},
		{
			name: "ошибка при обновлении статуса",
			input: &model.UserSetActive{
				UserID:   uuid.New(),
				IsActive: true,
			},
			setupMocks: func(repo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				dbError := errors.New("database connection error")

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(dbError)

				repo.On("SetActive", mock.Anything, mock.AnythingOfType("*model.UserSetActive")).Return(dbError)
			},
			expectedError: errors.New("database connection error"),
			checkResult: func(t *testing.T, u *model.User) {
				assert.Nil(t, u)
			},
		},
		{
			name: "ошибка при получении пользователя после обновления",
			input: &model.UserSetActive{
				UserID:   uuid.New(),
				IsActive: true,
			},
			setupMocks: func(repo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				getError := errors.New("failed to retrieve user")

				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(getError)

				repo.On("SetActive", mock.Anything, mock.AnythingOfType("*model.UserSetActive")).Return(nil)
				repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, getError)
			},
			expectedError: errors.New("failed to retrieve user"),
			checkResult: func(t *testing.T, u *model.User) {
				assert.Nil(t, u)
			},
		},
		{
			name: "нулевой UUID пользователя",
			input: &model.UserSetActive{
				UserID:   uuid.Nil,
				IsActive: true,
			},
			setupMocks: func(repo *mocks.MockUserRepository, txMgr *mocks.MockTxManager) {
				txMgr.On("ReadCommited", mock.Anything, mock.AnythingOfType("db.Handler")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(db.Handler)
						_ = fn(args.Get(0).(context.Context))
					}).Return(ErrNotFound)

				repo.On("SetActive", mock.Anything, mock.AnythingOfType("*model.UserSetActive")).Return(pgx.ErrNoRows)
			},
			expectedError: ErrNotFound,
			checkResult: func(t *testing.T, u *model.User) {
				assert.Nil(t, u)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockUserRepository(t)
			txMgr := mocks.NewMockTxManager(t)

			tt.setupMocks(repo, txMgr)

			svc := NewService(repo, txMgr)

			result, err := svc.SetActive(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedError, ErrNotFound) {
					assert.ErrorIs(t, err, ErrNotFound)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
