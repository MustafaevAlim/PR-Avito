package statistics

import (
	"net/http"

	"PR/internal/api/handlers"
	"PR/internal/service"
)

type StatisticsHandler struct {
	service service.StatisticsService
}

func NewHandler(service service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{service: service}
}

func mappingServiceError(err error) handlers.Error {
	var e handlers.Error
	switch err {

	default:
		e.Code = "UNKNOW"
		e.Message = err.Error()
		e.Status = http.StatusInternalServerError
	}
	return e
}
