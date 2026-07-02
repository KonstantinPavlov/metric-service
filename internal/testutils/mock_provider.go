package testutils


type MockProvider struct {
	CalledChan chan struct{}
}

func (m *MockProvider) GetCounters() map[string]int64 {
	m.CalledChan <- struct{}{}
	return map[string]int64{}
}
func (m *MockProvider) GetGauges() map[string]float64 {
	m.CalledChan <- struct{}{}
	return map[string]float64{}
}
func (m *MockProvider) SaveCounter(name string, value int64) error {
	m.CalledChan <- struct{}{}
	return nil
}
func (m *MockProvider) SaveGauge(name string, value float64) error {
	m.CalledChan <- struct{}{}
	return nil
}