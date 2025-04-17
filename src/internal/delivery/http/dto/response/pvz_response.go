package response

import (
	"avito-backend/src/internal/domain/models"
)

type GetPVZsResponse struct {
	Total int                         `json:"total"`
	PVZs  []*models.PVZWithReceptions `json:"pvzs"`
}
