package pr

import (
	"net/http"

	"PR/internal/api/handlers"
	"PR/internal/service"
	"PR/internal/service/pr"
)

type PullRequestHandler struct {
	service service.PullRequestService
}

func NewPullRequestHandler(serv service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{service: serv}
}

func mappingServiceError(err error) handlers.Error {
	var e handlers.Error
	switch err {
	case pr.ErrNotFound:
		e.Code = "NOT_FOUND"
		e.Message = "resource not found"
		e.Status = http.StatusBadRequest
	case pr.ErrPRExists:
		e.Code = "PR_EXISTS"
		e.Message = "PR id already exists"
		e.Status = http.StatusConflict
	case pr.ErrPRMerged:
		e.Code = "PR_MERGED"
		e.Message = "cannot reassign on merged PR"
		e.Status = http.StatusConflict
	case pr.ErrNoCandidate:
		e.Code = "NO_CANDIDATE"
		e.Message = "no active replacement candidate in team"
		e.Status = http.StatusConflict
	case pr.ErrNoAssigned:
		e.Code = "NOT_ASSIGNED"
		e.Message = "reviewer is not assigned to this PR"
		e.Status = http.StatusConflict

	default:
		e.Code = "UNKNOW"
		e.Message = err.Error()
		e.Status = http.StatusInternalServerError
	}
	return e
}
