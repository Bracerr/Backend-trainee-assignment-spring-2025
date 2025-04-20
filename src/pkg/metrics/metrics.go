package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP запросов",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Время ответа HTTP запросов",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 3, 5, 7, 10},
		},
		[]string{"method", "endpoint"},
	)

	PvzCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Общее количество созданных ПВЗ",
		},
	)

	ReceptionsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Общее количество созданных приёмок",
		},
	)

	ProductsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Общее количество добавленных товаров",
		},
	)
)
