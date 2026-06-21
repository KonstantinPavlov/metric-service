package repository

type MemStorage struct {
	Counters map[string]int64
	Gauges   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		make(map[string]int64),
		make(map[string]float64),
	}
}

func (storage *MemStorage) GetCounters() map[string]int64 {
	return storage.Counters
}
func (storage *MemStorage) GetGauges() map[string]float64 {
	return storage.Gauges
}

func (storage *MemStorage) SaveCounter(name string, value int64) error {
	storage.Counters[name] += value
	return nil
}

func (storage *MemStorage) SaveGauge(name string, value float64) error {
	storage.Gauges[name] = value
	return nil
}
