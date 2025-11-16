package team

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"PR/internal/api/handlers"
	"PR/internal/model"
)

func (h *TeamHandler) Create(c *gin.Context) {
	var t model.CreateTeamRequest

	err := c.ShouldBindJSON(&t)
	if err != nil {
		handlers.NewErrorResponse(c, handlers.BadRequestError())
		return
	}

	team := &model.Team{
		TeamName: t.TeamName,
		Members:  make([]*model.TeamMember, 0, len(t.Members)),
	}

	for _, m := range t.Members {
		var userID uuid.UUID
		var err error

		userID, err = uuid.Parse(m.ID)
		if err != nil {
			userID = handlers.StringToUUID(m.ID)
		}

		team.Members = append(team.Members, &model.TeamMember{
			ID:       userID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	err = h.service.Create(c.Request.Context(), team)
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": t,
	})
}
