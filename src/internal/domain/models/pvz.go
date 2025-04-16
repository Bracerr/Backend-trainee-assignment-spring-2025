package models

import (
	"time"

	"github.com/google/uuid"
)

type City string

const (
	Moscow City = "Москва"
	SPB    City = "Санкт-Петербург"
	Kazan  City = "Казань"
)

type PVZ struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             City      `json:"city"`
}

func (c City) IsValid() bool {
	return c == Moscow || c == SPB || c == Kazan
}
