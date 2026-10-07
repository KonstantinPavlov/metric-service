// Package handler реализует HTTP-обработчики для работы с метриками.
// Пакет поддерживает сбор, регистрацию наблюдателей для аудита и отображение
// накопленных метрик (типов counter и gauge) через HTML-шаблоны.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/KonstantinPavlov/metric-service/internal/audit"
	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// MetricHandler представляет собой структуру для обработки HTTP-запросов,
// связанных с метриками. Она управляет подписчиками (наблюдателями аудита)
// и взаимодействует с хранилищем метрик.
//
// Все методы структуры безопасны для конкурентного использования.
type MetricHandler struct {
	mu            sync.Mutex
	observers     []audit.AuditObserver
	Repository    repository.MetricRepository
	log           *zap.Logger
	wg            sync.WaitGroup
	eventsWorkers int
	eventsChannel chan audit.AuditEvent
}

// NewMetricHandler создает и инициализирует новый экземпляр MetricHandler
// на основе предоставленного репозитория и логгера.
func NewMetricHandler(repository repository.MetricRepository, log *zap.Logger, workers int) MetricHandler {
	return MetricHandler{
		observers:     make([]audit.AuditObserver, 0),
		Repository:    repository,
		log:           log,
		eventsWorkers: workers,
		eventsChannel: make(chan audit.AuditEvent, workers*2),
	}
}

// ListView представляет структуру данных метрики, подготовленную
// для вывода на веб-интерфейс или рендеринга в HTML-шаблонах.
type ListView struct {
	Name  string
	Type  string
	Value interface{}
}

// appenListView вспомогательная внутренняя функция для добавления метрики в срез отображения.
func appenListView(views []ListView, metricType string, metric repository.MetricData) []ListView {
	return append(views, ListView{
		Name:  metric.Name,
		Type:  metricType,
		Value: metric.Value},
	)
}

// Register регистрирует нового наблюдателя (AuditObserver) в обработчике.
// Наблюдатели используются для логирования действий или аудита изменений метрик.
func (mh *MetricHandler) Register(o audit.AuditObserver) {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	mh.observers = append(mh.observers, o)
}

func (mh *MetricHandler) Start(ctx context.Context) {
	for i := 0; i < mh.eventsWorkers; i++ {
		mh.wg.Add(1)
		go mh.worker(ctx)
	}
}

func (mh *MetricHandler) worker(ctx context.Context) {
	defer mh.wg.Done()
	for {
		select {
		case event, ok := <-mh.eventsChannel:
			if !ok {
				return
			}
			for _, o := range mh.observers {
				o.Notify(ctx, event)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (mh *MetricHandler) Stop() {
	mh.wg.Wait()
	for _, o := range mh.observers {
		o.Stop()
	}
}

// HandleList обрабатывает HTTP-запрос на получение полного списка метрик.
// Метод запрашивает все имена и значения метрик типа Counter и Gauge из репозитория,
// формирует срез структур ListView и передает его для рендеринга страницы "list-view.html".
//
// Возвращает ошибку, если репозиторию не удалось получить доступ к именам метрик,
// либо если произошел сбой при отрисовке HTML-шаблона. Ошибки чтения конкретных
// метрик пропускаются и пишутся в лог со статусом Error.
func (mh *MetricHandler) HandleList(c echo.Context) error {
	counterNames, err := mh.Repository.GetNames(c.Request().Context(), model.Counter)
	if err != nil {
		return err
	}
	counters := make([]repository.MetricData, 0)

	for _, counter := range counterNames {
		metric, err := mh.Repository.GetCounter(c.Request().Context(), counter)
		if err != nil {
			mh.log.Error("Failed to get counter!", zap.String("name", counter), zap.Error(err))
			continue
		}
		if metric != nil {
			counters = append(counters, *metric)
		}
	}
	gaugesNames, err := mh.Repository.GetNames(c.Request().Context(), model.Gauge)
	if err != nil {
		return err
	}
	gauges := make([]repository.MetricData, 0)

	for _, gauge := range gaugesNames {
		metric, err := mh.Repository.GetGauge(c.Request().Context(), gauge)
		if err != nil {
			mh.log.Error("Failed to get counter!", zap.String("name", gauge), zap.Error(err))
			continue
		}
		if metric != nil {
			gauges = append(gauges, *metric)
		}
	}

	metricsData := make([]ListView, 0)
	for _, metric := range counters {
		metricsData = appenListView(metricsData, model.Counter, metric)
	}

	for _, metric := range gauges {
		metricsData = appenListView(metricsData, model.Gauge, metric)
	}

	data := map[string]interface{}{
		"Metrics": metricsData,
	}

	return c.Render(http.StatusOK, "list-view.html", data)
}

// HandlePostValue обрабатывает HTTP-запрос POST для получения значения конкретной метрики.
// Метод принимает JSON-тело запроса, содержащее ID и тип метрики, извлекает её текущее
// значение из репозитория, подставляет значение в структуру ответа (Delta для counter, Value для gauge)
// и возвращает обновленный JSON-объект.
//
// Возвращаемые HTTP-статусы:
//   - 200 OK: Метрика успешно найдена и возвращена в формате JSON.
//   - 400 Bad Request: Некорректный формат JSON или передан неизвестный тип метрики.
//   - 404 Not Found: Не указано имя метрики или метрика с таким ID отсутствует в репозитории.
//   - 500 Internal Server Error: Произошла внутренняя ошибка репозитория при извлечении данных.
func (mh *MetricHandler) HandlePostValue(c echo.Context) error {
	req := &model.Metrics{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("%v", err))
	}

	if req.ID == "" {
		return c.String(http.StatusNotFound, "Metric name must be set")
	}

	switch req.MType {
	case model.Counter:
		metric, err := mh.Repository.GetCounter(c.Request().Context(), req.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}
		if metric == nil {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		metricValue := metric.Value.(int64)
		req.Delta = &metricValue
		return c.JSON(http.StatusOK, req)
	case model.Gauge:
		metric, err := mh.Repository.GetGauge(c.Request().Context(), req.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}
		if metric == nil {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		metricValue := metric.Value.(float64)
		req.Value = &metricValue
		return c.JSON(http.StatusOK, req)
	default:
		return c.String(http.StatusBadRequest, "unkwnown metric type!")
	}
}

// HandleGetValue обрабатывает HTTP-запрос GET для получения значения метрики через URL-параметры.
// Тип метрики и её имя считываются из параметров пути "type" и "name". Значение возвращается
// в виде простой строки (plain text).
//
// Возвращаемые HTTP-статусы:
//   - 200 OK: Значение метрики успешно возвращено в текстовом виде.
//   - 400 Bad Request: Передан неизвестный тип метрики.
//   - 404 Not Found: Не указано имя метрики или метрика отсутствует в репозитории.
//   - 500 Internal Server Error: Ошибка репозитория или несоответствие типов данных при приведении.
func (mh *MetricHandler) HandleGetValue(c echo.Context) error {
	metricType := c.Param("type")
	metricName := c.Param("name")

	if metricName == "" {
		return c.String(http.StatusNotFound, "Metric name must be set")
	}

	switch metricType {
	case model.Counter:
		metric, err := mh.Repository.GetCounter(c.Request().Context(), metricName)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}
		if metric == nil {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		val, ok := metric.Value.(int64)
		if !ok {
			return c.String(http.StatusInternalServerError, "invalid counter type")
		}
		return c.String(http.StatusOK, strconv.FormatInt(val, 10))

	case model.Gauge:
		metric, err := mh.Repository.GetGauge(c.Request().Context(), metricName)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}
		if metric == nil {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		val, ok := metric.Value.(float64)
		if !ok {
			return c.String(http.StatusInternalServerError, "invalid gauge type")
		}
		return c.String(http.StatusOK, strconv.FormatFloat(val, 'f', -1, 64))
	default:
		return c.String(http.StatusBadRequest, "unkwnown metric type!")
	}
}

// HandleBodyUpdate обрабатывает HTTP-запрос POST для обновления или создания метрики.
// Данные считываются из JSON-тела запроса. После успешного сохранения метрики в хранилище,
// метод уведомляет всех зарегистрированных наблюдателей (observers) о событии обновления.
//
// Возвращаемые HTTP-статусы:
//   - 200 OK: Метрика успешно обновлена, тело запроса возвращено обратно в формате JSON.
//   - 400 Bad Request: Некорректный JSON, отсутствует обязательное поле (Delta или Value)
//     или передан неизвестный тип метрики.
//   - 404 Not Found: Не указан ID метрики.
//   - 500 Internal Server Error: Ошибка репозитория при попытке сохранить данные.
func (mh *MetricHandler) HandleBodyUpdate(c echo.Context) error {
	req := &model.Metrics{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("%v", err))
	}

	if req.ID == "" {
		return c.String(http.StatusNotFound, "Metric name must be set")
	}

	switch req.MType {
	case model.Counter:
		if req.Delta == nil {
			return c.String(http.StatusBadRequest, "delta not specified!")
		}
		mh.log.Info("Saving data for counter metric", zap.String("metric", req.ID))
		err = mh.Repository.SaveCounter(c.Request().Context(), req.ID, *req.Delta)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save counter: %v", err))
		}
	case model.Gauge:
		if req.Value == nil {
			return c.String(http.StatusBadRequest, "value not specified!")
		}
		err = mh.Repository.SaveGauge(c.Request().Context(), req.ID, *req.Value)
		mh.log.Info("Saving data for gauge metric", zap.String("metric", req.ID))
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save counter: %v", err))
		}
	default:
		return c.String(http.StatusBadRequest, "unkwnown metric type!")
	}

	mh.notify(
		c.Request().Context(),
		audit.NewServerEvent([]string{req.ID}, c.RealIP()),
	)
	return c.JSON(http.StatusOK, req)
}

// HandleParamUpdate обрабатывает HTTP-запрос POST для обновления метрики через URL-параметры пути.
// Параметры "type", "name" и "value" считываются непосредственно из URL. Метод производит
// валидацию и парсинг строкового значения в числовой тип данных, сохраняет метрику
// и оповещает зарегистрированных наблюдателей.
//
// Возвращаемые HTTP-статусы:
//   - 200 OK: Метрика успешно сохранена, возвращается текстовое сообщение "metric saved".
//   - 400 Bad Request: Не удалось распарсить числовое значение или указан неизвестный тип метрики.
//   - 404 Not Found: Не указано имя метрики.
//   - 500 Internal Server Error: Ошибка репозитория при сохранении.
func (mh *MetricHandler) HandleParamUpdate(c echo.Context) error {
	metricType := c.Param("type")
	metricName := c.Param("name")
	value := c.Param("value")

	if metricName == "" {
		return c.String(http.StatusNotFound, "Metric name must be set")
	}

	switch metricType {
	case model.Counter:
		vInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "value must be a number")
		}
		mh.log.Info("Saving data for counter metric", zap.String("metric", metricName))
		err = mh.Repository.SaveCounter(c.Request().Context(), metricName, vInt)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save counter: %v", err))
		}
	case model.Gauge:
		vFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Failed to parse gauge value: %v", err))
		}
		err = mh.Repository.SaveGauge(c.Request().Context(), metricName, vFloat)
		mh.log.Info("Saving data for gauge metric", zap.String("metric", metricName))
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save counter: %v", err))
		}
	default:
		return c.String(http.StatusBadRequest, "unkwnown metric type!")
	}
	mh.notify(
		c.Request().Context(),
		audit.NewServerEvent([]string{metricName}, c.Request().Host),
	)
	return c.String(http.StatusOK, "metric saved")
}

// HandleUpdates обрабатывает HTTP-запрос POST для пакетного (batch) обновления метрик.
// Метод принимает JSON-массив структур model.Metrics, валидирует переданные элементы,
// группирует их по типам и отправляет в репозиторий для атомарного сохранения в рамках транзакции.
// После успешного сохранения пакета метод отправляет одно общее уведомление для всех измененных метрик.
//
// Возвращаемые HTTP-статусы:
//   - 200 OK: Пакет метрик успешно сохранен, массив возвращается в формате JSON.
//   - 400 Bad Request: Пустой массив данных, ошибка десериализации JSON,
//     отсутствие дельты/значения у метрики или неизвестный тип метрики в пакете.
//   - 500 Internal Server Error: Сбой транзакции или пакетного сохранения в репозитории.
func (mh *MetricHandler) HandleUpdates(c echo.Context) error {
	var requests []model.Metrics
	err := c.Bind(&requests)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("%v", err))
	}

	if len(requests) == 0 {
		return c.String(http.StatusBadRequest, "Received empty payload!")
	}

	var countersData []repository.MetricData
	var gaugesData []repository.MetricData
	var metricNames []string
	// Валидируем и распределяем метрики по типам
	for _, metric := range requests {
		switch metric.MType {
		case model.Counter:
			if metric.Delta == nil {
				return c.String(http.StatusBadRequest, fmt.Sprintf("Delta not specified for counter: %s", metric.ID))
			}
			metricNames = append(metricNames, metric.ID)
			countersData = append(countersData, repository.MetricData{
				Name:  metric.ID,
				Value: *metric.Delta,
			})
		case model.Gauge:
			if metric.Value == nil {
				return c.String(http.StatusBadRequest, fmt.Sprintf("Value not specified for gauge: %s", metric.ID))
			}
			metricNames = append(metricNames, metric.ID)
			gaugesData = append(gaugesData, repository.MetricData{
				Name:  metric.ID,
				Value: *metric.Value,
			})
		default:
			return c.String(http.StatusBadRequest, fmt.Sprintf("Unknown metric type: %s", metric.MType))
		}
	}

	mh.log.Info("Saving metrics in transaction",
		zap.Int("counters", len(countersData)),
		zap.Int("gauges", len(gaugesData)),
	)

	err = mh.Repository.SaveMetrics(c.Request().Context(), countersData, gaugesData)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save metrics: %v", err))
	}
	mh.notify(
		c.Request().Context(),
		audit.NewServerEvent(metricNames, c.Request().Host),
	)
	return c.JSON(http.StatusOK, requests)
}

// notify добавляет в канал событие аудита
func (mh *MetricHandler) notify(ctx context.Context, e audit.AuditEvent) {
	select {
	case mh.eventsChannel <- e:
	case <-ctx.Done():
		return
	}
}
