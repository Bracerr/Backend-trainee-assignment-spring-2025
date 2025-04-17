package request

type CreatePVZRequest struct {
	City string `json:"city"`
}

type GetPVZsRequest struct {
	StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02"`
	Offset    int    `json:"offset" validate:"required,min=0"`
	Limit     int    `json:"limit" validate:"required,min=1,max=100"`
}
