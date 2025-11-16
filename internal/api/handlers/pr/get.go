package pr

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"PR/internal/api/handlers"
)

func (h *PullRequestHandler) GetByReviewer(c *gin.Context) {
	userID := c.Query("user_id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		userUUID = handlers.StringToUUID(userID)
	}

	prs, err := h.service.GetByReviewer(c.Request.Context(), userUUID)
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})

}
