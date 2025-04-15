package models

type Role string

const (
	EmployeeRole  Role = "employee"
	ModeratorRole Role = "moderator"
)

type User struct {
	Role Role `json:"role"`
}
