package repository

type MetricRepository interface {
	GetCounters() map[string]int64
	GetGauges() map[string]float64
	SaveCounter(name string, value int64) error
	SaveGauge(name string, value float64) error
}