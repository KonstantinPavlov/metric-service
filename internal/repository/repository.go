package repository

type MetricRepository interface {
	GetNames(metricType string) []string	
	GetCounter(name string) *MetricData
	GetGauge(name string) *MetricData
	SaveCounter(name string, value int64) error
	SaveGauge(name string, value float64) error
}

type MetricData struct {
	Name  string
	Value interface{}
}
