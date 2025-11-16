package pr

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"PR/internal/api/handlers"
	"PR/internal/model"
)

func (h *PullRequestHandler) Merge(c *gin.Context) {
	var p model.PullRequestInMerge

	err := c.BindJSON(&p)
	if err != nil {
		handlers.NewErrorResponse(c, handlers.BadRequestError())
		return
	}

	id, err := uuid.Parse(p.ID)
	if err != nil {
		id = handlers.StringToUUID(p.ID)
	}

	pr, err := h.service.Merge(c.Request.Context(), id)
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr": pr,
	})
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var req model.PullRequestInReassign

	err := c.BindJSON(&req)
	if err != nil {
		handlers.NewErrorResponse(c, handlers.BadRequestError())
		return
	}

	oldID, err := uuid.Parse(req.OldReviewerID)
	if err != nil {
		oldID = handlers.StringToUUID(req.OldReviewerID)
	}

	prID, err := uuid.Parse(req.PrID)
	if err != nil {
		prID = handlers.StringToUUID(req.PrID)
	}

	pr, replacedBy, err := h.service.ReassignReviewers(c.Request.Context(), oldID, prID)
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": replacedBy,
	})

}
