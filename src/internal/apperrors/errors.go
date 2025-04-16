package apperrors

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRole        = errors.New("invalid role")
	ErrValidationFailed   = errors.New("validation failed")
	ErrInvalidCity        = errors.New("invalid city")
)
