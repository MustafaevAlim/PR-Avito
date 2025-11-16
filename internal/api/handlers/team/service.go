package team

import (
	"net/http"

	"PR/internal/api/handlers"
	"PR/internal/service"
	"PR/internal/service/team"
)

type TeamHandler struct {
	service service.TeamService
}

func NewTeamHandler(serv service.TeamService) *TeamHandler {
	return &TeamHandler{service: serv}
}

func mappingServiceError(err error) handlers.Error {
	var e handlers.Error
	switch err {
	case team.ErrTeamExist:
		e.Code = "TEAM_EXISTS"
		e.Message = "team_name already exists"
		e.Status = http.StatusBadRequest
	default:
		e.Code = "UNKNOW"
		e.Message = err.Error()
		e.Status = http.StatusInternalServerError
	}
	return e
}
