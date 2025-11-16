package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Err Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func NewErrorResponse(c *gin.Context, err Error) {
	c.AbortWithStatusJSON(err.Status, ErrorResponse{Err: err})
}

func BadRequestError() Error {
	return Error{
		Code:    "BAD_REQUEST",
		Message: "invalid data",
		Status:  http.StatusBadRequest,
	}
}
