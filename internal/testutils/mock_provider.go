package testutils

import "context"

type MockProvider struct {
	CalledChan chan struct{}
}

func (m *MockProvider) GetCounters(ctx context.Context) map[string]int64 {
	m.CalledChan <- struct{}{}
	return map[string]int64{}
}
func (m *MockProvider) GetGauges(ctx context.Context) map[string]float64 {
	m.CalledChan <- struct{}{}
	return map[string]float64{}
}
func (m *MockProvider) SaveCounter(ctx context.Context, name string, value int64) error {
	m.CalledChan <- struct{}{}
	return nil
}
func (m *MockProvider) SaveGauge(ctx context.Context, name string, value float64) error {
	m.CalledChan <- struct{}{}
	return nil
}
