package repository

type MetricRepository interface {
	GetNames(metricType string) ([]string, error)
	GetCounter(name string) (*MetricData, error)
	GetGauge(name string) (*MetricData, error)
	SaveCounter(name string, value int64) error
	SaveGauge(name string, value float64) error
	Ping() error
	Start() error
	Stop()
}

type MetricData struct {
	Name  string
	Value interface{}
}
