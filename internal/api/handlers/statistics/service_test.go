package statistics_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"PR/internal/api/handlers/statistics"
	"PR/internal/mocks"
	"PR/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetReviewerStats(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStatisticsService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			setupMock: func(m *mocks.MockStatisticsService) {
				m.On("GetReviewerStatistics", mock.Anything).
					Return([]*model.ReviewerStats{
						{
							ReviewerID:    uuid.New(),
							ReviewerName:  "John Doe",
							AssignedCount: 5,
						},
						{
							ReviewerID:    uuid.New(),
							ReviewerName:  "Jane Smith",
							AssignedCount: 3,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats []model.ReviewerStats
				err := json.Unmarshal(w.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Len(t, stats, 2)
				assert.Equal(t, "John Doe", stats[0].ReviewerName)
			},
		},
		{
			name: "service_error",
			setupMock: func(m *mocks.MockStatisticsService) {
				m.On("GetReviewerStatistics", mock.Anything).
					Return(([]*model.ReviewerStats)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name: "empty_result",
			setupMock: func(m *mocks.MockStatisticsService) {
				m.On("GetReviewerStatistics", mock.Anything).
					Return([]*model.ReviewerStats{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats []model.ReviewerStats
				err := json.Unmarshal(w.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Len(t, stats, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := mocks.NewMockStatisticsService(t)
			tt.setupMock(mockService)

			handler := statistics.NewHandler(mockService)
			router.GET("/statistics/reviewers", handler.GetReviewerStats)

			req, _ := http.NewRequest("GET", "/statistics/reviewers", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetPRStats(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStatisticsService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			setupMock: func(m *mocks.MockStatisticsService) {
				m.On("GetPRStatistics", mock.Anything).
					Return([]*model.PRStats{
						{
							PRID:          uuid.New(),
							PRName:        "Feature XYZ",
							ReviewerCount: 3,
							Status:        "open",
						},
						{
							PRID:          uuid.New(),
							PRName:        "Bugfix ABC",
							ReviewerCount: 2,
							Status:        "merged",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats []model.PRStats
				err := json.Unmarshal(w.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Len(t, stats, 2)
				assert.Equal(t, "Feature XYZ", stats[0].PRName)
				assert.Equal(t, 3, stats[0].ReviewerCount)
			},
		},
		{
			name: "service_error",
			setupMock: func(m *mocks.MockStatisticsService) {
				m.On("GetPRStatistics", mock.Anything).
					Return(([]*model.PRStats)(nil), errors.New("connection timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name: "empty_result",
			setupMock: func(m *mocks.MockStatisticsService) {
				m.On("GetPRStatistics", mock.Anything).
					Return([]*model.PRStats{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats []model.PRStats
				err := json.Unmarshal(w.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Len(t, stats, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := mocks.NewMockStatisticsService(t)
			tt.setupMock(mockService)

			handler := statistics.NewHandler(mockService)
			router.GET("/statistics/prs", handler.GetPRStats)

			req, _ := http.NewRequest("GET", "/statistics/prs", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockService.AssertExpectations(t)
		})
	}
}
