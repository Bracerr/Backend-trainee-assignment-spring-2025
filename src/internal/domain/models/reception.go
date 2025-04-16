package models

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionStatus string

const (
	InProgress ReceptionStatus = "in_progress"
	Closed     ReceptionStatus = "close"
)

type Reception struct {
	ID       uuid.UUID       `json:"id"`
	DateTime time.Time       `json:"dateTime"`
	PVZID    uuid.UUID       `json:"pvzId"`
	Status   ReceptionStatus `json:"status"`
	Products []Product       `json:"products"`
}
