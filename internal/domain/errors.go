package domain

import "errors"

var (
	ErrInternal         = errors.New("internal server error")
	ErrNotFound         = errors.New("resource not found")
	ErrUserNotConnected = errors.New("user not connected")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
)
