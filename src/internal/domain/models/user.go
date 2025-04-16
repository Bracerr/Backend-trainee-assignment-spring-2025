package models

import "github.com/google/uuid"

type Role string

const (
	EmployeeRole  Role = "employee"
	ModeratorRole Role = "moderator"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"-"`
}
