package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidUserRequest = errors.New("invalid user request")
	ErrInvalidUserID      = errors.New("invalid user ID")
)
