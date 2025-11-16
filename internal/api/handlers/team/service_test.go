package team_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"PR/internal/api/handlers/team"
	"PR/internal/mocks"
	"PR/internal/model"
	serviceTeam "PR/internal/service/team"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      interface{}
		setupMock      func(*mocks.MockTeamService)
		expectedStatus int
	}{
		{
			name: "success",
			inputBody: model.CreateTeamRequest{
				TeamName: "Backend Team",
				Members: []model.MemberRequest{
					{
						ID:       uuid.New().String(),
						Username: "john_doe",
						IsActive: true,
					},
					{
						ID:       uuid.New().String(),
						Username: "jane_smith",
						IsActive: true,
					},
				},
			},
			setupMock: func(m *mocks.MockTeamService) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(team *model.Team) bool {
					return team.TeamName == "Backend Team" && len(team.Members) == 2
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "team_already_exists",
			inputBody: model.CreateTeamRequest{
				TeamName: "Existing Team",
				Members: []model.MemberRequest{
					{
						ID:       uuid.New().String(),
						Username: "user1",
						IsActive: true,
					},
				},
			},
			setupMock: func(m *mocks.MockTeamService) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(serviceTeam.ErrTeamExist)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_json",
			inputBody:      "invalid json",
			setupMock:      func(m *mocks.MockTeamService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty_team_name",
			inputBody: model.CreateTeamRequest{
				TeamName: "",
				Members: []model.MemberRequest{
					{
						ID:       uuid.New().String(),
						Username: "user1",
						IsActive: true,
					},
				},
			},
			setupMock: func(m *mocks.MockTeamService) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("team name is required"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			router := gin.New()

			mockService := mocks.NewMockTeamService(t)
			tt.setupMock(mockService)

			handler := team.NewTeamHandler(mockService)
			router.POST("/team", handler.Create)

			var body []byte
			if str, ok := tt.inputBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.inputBody)
			}

			req, _ := http.NewRequest("POST", "/team", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetTeamByName(t *testing.T) {
	tests := []struct {
		name           string
		teamName       string
		setupMock      func(*mocks.MockTeamService, string)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "success",
			teamName: "Backend Team",
			setupMock: func(m *mocks.MockTeamService, name string) {
				m.On("GetTeamByName", mock.Anything, name).
					Return(&model.Team{
						TeamName: "Backend Team",
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
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var team model.Team
				err := json.Unmarshal(w.Body.Bytes(), &team)
				assert.NoError(t, err)
				assert.Equal(t, "Backend Team", team.TeamName)
				assert.Len(t, team.Members, 2)
			},
		},
		{
			name:     "team_not_found",
			teamName: "NonExistent Team",
			setupMock: func(m *mocks.MockTeamService, name string) {
				m.On("GetTeamByName", mock.Anything, name).
					Return((*model.Team)(nil), errors.New("team not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name:     "empty_team_name",
			teamName: "",
			setupMock: func(m *mocks.MockTeamService, name string) {
				m.On("GetTeamByName", mock.Anything, name).
					Return((*model.Team)(nil), errors.New("team name is required"))
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

			mockService := mocks.NewMockTeamService(t)
			tt.setupMock(mockService, tt.teamName)

			handler := team.NewTeamHandler(mockService)
			router.GET("/team", handler.GetTeamByName)

			req, _ := http.NewRequest("GET", "/team?team_name="+tt.teamName, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockService.AssertExpectations(t)
		})
	}
}
