package team

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"PR/internal/api/handlers"
)

func (h *TeamHandler) GetTeamByName(c *gin.Context) {
	name := c.Query("team_name")

	t, err := h.service.GetTeamByName(c.Request.Context(), name)
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusOK, t)
}
