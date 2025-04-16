package apperrors

import "errors"

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidRole           = errors.New("invalid role")
	ErrValidationFailed      = errors.New("validation failed")
	ErrInvalidCity           = errors.New("invalid city")
	ErrActiveReceptionExists = errors.New("active reception exists")
	ErrNoActiveReception     = errors.New("no active reception")
	ErrInvalidProductType    = errors.New("invalid product type")
	ErrPVZNotFound           = errors.New("pvz not found")
)
