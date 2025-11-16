package user

import (
	"net/http"

	"PR/internal/api/handlers"
	"PR/internal/service"
	"PR/internal/service/user"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(serv service.UserService) *UserHandler {
	return &UserHandler{service: serv}
}

func mappingServiceError(err error) handlers.Error {
	var e handlers.Error
	switch err {
	case user.ErrNotFound:
		e.Code = "NOT_FOUND"
		e.Message = "resource not found"
		e.Status = http.StatusNotFound
	default:
		e.Code = "UNKNOW"
		e.Message = err.Error()
		e.Status = http.StatusInternalServerError
	}
	return e
}
