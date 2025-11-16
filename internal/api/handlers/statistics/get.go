package statistics

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"PR/internal/api/handlers"
)

func (h *StatisticsHandler) GetReviewerStats(c *gin.Context) {
	stats, err := h.service.GetReviewerStatistics(c.Request.Context())
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *StatisticsHandler) GetPRStats(c *gin.Context) {
	stats, err := h.service.GetPRStatistics(c.Request.Context())
	if err != nil {
		handlers.NewErrorResponse(c, mappingServiceError(err))
		return
	}

	c.JSON(http.StatusOK, stats)
}
