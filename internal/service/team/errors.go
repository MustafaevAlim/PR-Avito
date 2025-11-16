package team

import "errors"

var (
	ErrTeamExist = errors.New("team exist")
	ErrNotFound  = errors.New("team not found")
)
