package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductType string

const (
	Electronics ProductType = "электроника"
	Clothes     ProductType = "одежда"
	Shoes       ProductType = "обувь"
)

type Product struct {
	ID          uuid.UUID   `json:"id"`
	DateTime    time.Time   `json:"dateTime"`
	Type        ProductType `json:"type"`
	ReceptionID uuid.UUID   `json:"receptionId"`
}

func (t ProductType) IsValid() bool {
	switch t {
	case Electronics, Clothes, Shoes:
		return true
	}
	return false
}
