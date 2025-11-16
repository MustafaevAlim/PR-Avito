package pr_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"PR/internal/api/handlers/pr"
	"PR/internal/mocks"
	"PR/internal/model"
	servicePr "PR/internal/service/pr"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      interface{}
		setupMock      func(*mocks.MockPullRequestService)
		expectedStatus int
	}{
		{
			name: "success",
			inputBody: model.PullRequestInCreate{
				ID:       uuid.New().String(),
				Name:     "Test PR",
				AuthorID: uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(&model.PullRequest{
						ID:       uuid.New(),
						Name:     "Test PR",
						AuthorID: uuid.New(),
						Status:   "open",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "pr_already_exists",
			inputBody: model.PullRequestInCreate{
				ID:       uuid.New().String(),
				Name:     "Existing PR",
				AuthorID: uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService) {
				m.On("Create", mock.Anything, mock.Anything).
					Return((*model.PullRequest)(nil), servicePr.ErrPRExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "invalid_json",
			inputBody:      "invalid json",
			setupMock:      func(m *mocks.MockPullRequestService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := new(mocks.MockPullRequestService)
			tt.setupMock(mockService)

			handler := pr.NewPullRequestHandler(mockService)
			router.POST("/pr", handler.Create)

			var body []byte
			if str, ok := tt.inputBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.inputBody)
			}

			req, _ := http.NewRequest("POST", "/pr", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetByReviewer(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*mocks.MockPullRequestService, uuid.UUID)
		expectedStatus int
	}{
		{
			name:   "success",
			userID: uuid.New().String(),
			setupMock: func(m *mocks.MockPullRequestService, id uuid.UUID) {
				m.On("GetByReviewer", mock.Anything, id).
					Return([]*model.PullRequestShort{
						{ID: uuid.New(), Name: "PR 1", AuthorID: uuid.New()},
						{ID: uuid.New(), Name: "PR 2", AuthorID: uuid.New()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "not_found",
			userID: uuid.New().String(),
			setupMock: func(m *mocks.MockPullRequestService, id uuid.UUID) {
				m.On("GetByReviewer", mock.Anything, id).
					Return(([]*model.PullRequestShort)(nil), servicePr.ErrNotFound)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := new(mocks.MockPullRequestService)
			userUUID, _ := uuid.Parse(tt.userID)
			tt.setupMock(mockService, userUUID)

			handler := pr.NewPullRequestHandler(mockService)
			router.GET("/pr/reviewer", handler.GetByReviewer)

			req, _ := http.NewRequest("GET", "/pr/reviewer?user_id="+tt.userID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      model.PullRequestInMerge
		setupMock      func(*mocks.MockPullRequestService, uuid.UUID)
		expectedStatus int
	}{
		{
			name: "success",
			inputBody: model.PullRequestInMerge{
				ID: uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService, id uuid.UUID) {
				m.On("Merge", mock.Anything, id).
					Return(&model.PullRequest{
						ID:       id,
						Name:     "Merged PR",
						AuthorID: uuid.New(),
						Status:   "merged",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "already_merged",
			inputBody: model.PullRequestInMerge{
				ID: uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService, id uuid.UUID) {
				m.On("Merge", mock.Anything, id).
					Return((*model.PullRequest)(nil), servicePr.ErrPRMerged)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "not_found",
			inputBody: model.PullRequestInMerge{
				ID: uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService, id uuid.UUID) {
				m.On("Merge", mock.Anything, id).
					Return((*model.PullRequest)(nil), servicePr.ErrNotFound)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := new(mocks.MockPullRequestService)
			prID, _ := uuid.Parse(tt.inputBody.ID)
			tt.setupMock(mockService, prID)

			handler := pr.NewPullRequestHandler(mockService)
			router.POST("/pr/merge", handler.Merge)

			body, _ := json.Marshal(tt.inputBody)
			req, _ := http.NewRequest("POST", "/pr/merge", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestReassign(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      model.PullRequestInReassign
		setupMock      func(*mocks.MockPullRequestService, uuid.UUID, uuid.UUID)
		expectedStatus int
	}{
		{
			name: "success",
			inputBody: model.PullRequestInReassign{
				OldReviewerID: uuid.New().String(),
				PrID:          uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService, oldID, prID uuid.UUID) {
				m.On("ReassignReviewers", mock.Anything, oldID, prID).
					Return(&model.PullRequest{
						ID:       prID,
						Name:     "Test PR",
						AuthorID: uuid.New(),
					}, uuid.New(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "no_candidate",
			inputBody: model.PullRequestInReassign{
				OldReviewerID: uuid.New().String(),
				PrID:          uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService, oldID, prID uuid.UUID) {
				m.On("ReassignReviewers", mock.Anything, oldID, prID).
					Return((*model.PullRequest)(nil), uuid.Nil, servicePr.ErrNoCandidate)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "not_assigned",
			inputBody: model.PullRequestInReassign{
				OldReviewerID: uuid.New().String(),
				PrID:          uuid.New().String(),
			},
			setupMock: func(m *mocks.MockPullRequestService, oldID, prID uuid.UUID) {
				m.On("ReassignReviewers", mock.Anything, oldID, prID).
					Return((*model.PullRequest)(nil), uuid.Nil, servicePr.ErrNoAssigned)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := new(mocks.MockPullRequestService)
			oldID, _ := uuid.Parse(tt.inputBody.OldReviewerID)
			prID, _ := uuid.Parse(tt.inputBody.PrID)
			tt.setupMock(mockService, oldID, prID)

			handler := pr.NewPullRequestHandler(mockService)
			router.POST("/pr/reassign", handler.Reassign)

			body, _ := json.Marshal(tt.inputBody)
			req, _ := http.NewRequest("POST", "/pr/reassign", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
