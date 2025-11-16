package pr

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"PR/internal/api/handlers"
	"PR/internal/model"
)

func (h *PullRequestHandler) Create(c *gin.Context) {
	var pr model.PullRequestInCreate

	err := c.BindJSON(&pr)
	if err != nil {
		handlers.NewErrorResponse(c, handlers.BadRequestError())
		return
	}
	prID, err := uuid.Parse(pr.ID)
	if err != nil {
		prID = handlers.StringToUUID(pr.ID)
	}

	prAuthorID, err := uuid.Parse(pr.AuthorID)
	if err != nil {
		prAuthorID = handlers.StringToUUID(pr.AuthorID)
	}

	prReturning, err := h.service.Create(c.Request.Context(), &model.PullRequestShort{
		ID:       prID,
		Name:     pr.Name,
		AuthorID: prAuthorID,
	})
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusCreated, prReturning)
}
