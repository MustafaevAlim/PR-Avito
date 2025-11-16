package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"PR/internal/api/handlers"
	"PR/internal/model"
)

func (h *UserHandler) SetActive(c *gin.Context) {
	var req model.UserSetActiveRequest

	err := c.BindJSON(&req)
	if err != nil {
		handlers.NewErrorResponse(c, handlers.BadRequestError())
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		userID = handlers.StringToUUID(req.UserID)
	}

	u, err := h.service.SetActive(c.Request.Context(), &model.UserSetActive{
		UserID:   userID,
		IsActive: req.IsActive,
	})
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": u,
	})

}
