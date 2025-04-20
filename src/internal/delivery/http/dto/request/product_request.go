package request

type CreateProductRequest struct {
	Type  string `json:"type"`
	PVZID string `json:"pvzId"`
}
