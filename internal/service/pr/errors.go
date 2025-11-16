package pr

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrNotActive   = errors.New("not active")
	ErrPRExists    = errors.New("PR exists")
	ErrNoCandidate = errors.New("no candidate")
	ErrPRMerged    = errors.New("PR merged")
	ErrNoAssigned  = errors.New("no assigned")
)
