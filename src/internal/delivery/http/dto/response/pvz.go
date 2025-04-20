package response

import "time"

type PVZWithReceptionsResponse struct {
	PVZ struct {
		ID               string    `json:"id"`
		RegistrationDate time.Time `json:"registrationDate"`
		City             string    `json:"city"`
	} `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

type ReceptionWithProducts struct {
	Reception struct {
		ID       string    `json:"id"`
		DateTime time.Time `json:"dateTime"`
		PVZID    string    `json:"pvzId"`
		Status   string    `json:"status"`
	} `json:"reception"`
	Products []Product `json:"products"`
}

type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"receptionId"`
} 