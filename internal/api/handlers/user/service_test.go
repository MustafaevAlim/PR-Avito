package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"PR/internal/api/handlers/user"
	"PR/internal/mocks"
	"PR/internal/model"
	serviceUser "PR/internal/service/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetActive(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      interface{}
		setupMock      func(*mocks.MockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "activate_user",
			inputBody: model.UserSetActiveRequest{
				UserID:   uuid.New().String(),
				IsActive: true,
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("SetActive", mock.Anything, mock.MatchedBy(func(req *model.UserSetActive) bool {
					return req.IsActive == true
				})).Return(&model.User{
					ID:       uuid.New(),
					Username: "john_doe",
					TeamName: "Backend Team",
					IsActive: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]model.User
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["user"].IsActive)
			},
		},
		{
			name: "deactivate_user",
			inputBody: model.UserSetActiveRequest{
				UserID:   uuid.New().String(),
				IsActive: false,
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("SetActive", mock.Anything, mock.MatchedBy(func(req *model.UserSetActive) bool {
					return req.IsActive == false
				})).Return(&model.User{
					ID:       uuid.New(),
					Username: "jane_smith",
					TeamName: "Frontend Team",
					IsActive: false,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]model.User
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["user"].IsActive)
			},
		},
		{
			name: "user_not_found",
			inputBody: model.UserSetActiveRequest{
				UserID:   uuid.New().String(),
				IsActive: true,
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("SetActive", mock.Anything, mock.Anything).
					Return((*model.User)(nil), serviceUser.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name:           "invalid_json",
			inputBody:      "invalid json",
			setupMock:      func(m *mocks.MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name: "service_error",
			inputBody: model.UserSetActiveRequest{
				UserID:   uuid.New().String(),
				IsActive: true,
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("SetActive", mock.Anything, mock.Anything).
					Return((*model.User)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := mocks.NewMockUserService(t)
			tt.setupMock(mockService)

			handler := user.NewUserHandler(mockService)
			router.POST("/user/active", handler.SetActive)

			var body []byte
			if str, ok := tt.inputBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.inputBody)
			}

			req, _ := http.NewRequest("POST", "/user/active", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockService.AssertExpectations(t)
		})
	}
}
