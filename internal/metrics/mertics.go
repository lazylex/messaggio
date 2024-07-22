package metrics

import (
	"errors"
	"fmt"
	"github.com/lazylex/messaggio/internal/config"

	httpMetrics "github.com/lazylex/messaggio/internal/ports/metrics/http"
	"github.com/lazylex/messaggio/internal/ports/metrics/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
)

const (
	NAMESPACE = "messaggio"
	PATH      = "path"
)

// Metrics структура, содержащая объекты, реализующие интерфейсы для сбора метрик.
type Metrics struct {
	HTTP    httpMetrics.MetricsInterface
	Service service.MetricsInterface
}

// MustCreate возвращает метрики *Metrics или останавливает программу, если не удалось запустить http сервер для
// работы с Prometheus или занести метрики в регистр.
func MustCreate(cfg *config.Prometheus) *Metrics {
	var port = "9323"
	var url = "/metrics"

	if len(cfg.PrometheusPort) > 0 {
		port = cfg.PrometheusPort
	}

	if len(cfg.PrometheusMetricsURL) > 0 {
		url = cfg.PrometheusMetricsURL
	}

	startHTTP(url, port)

	metrics, err := registerMetrics()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return metrics
}

// registerMetrics заносит метрики в регистр и возвращает их. При неудаче возвращает ошибку.
func registerMetrics() (*Metrics, error) {
	var err error
	var incomingMsgMetric, processedMetric, problemsSavingMetric, requests *prometheus.CounterVec
	var requestDuration *prometheus.HistogramVec

	if requests, err = createHTTPRequestsTotalMetric(); err != nil {
		return nil, err
	}

	if requestDuration, err = createHTTPRequestDurationSecondsBucketMetric(); err != nil {
		return nil, err
	}

	if incomingMsgMetric, err = createIncomingMsgTotalMetric(); err != nil {
		return nil, err
	}

	if processedMetric, err = createProcessedMsgTotalMetric(); err != nil {
		return nil, err
	}

	if problemsSavingMetric, err = createProblemsSavingInDBTotalMetric(); err != nil {
		return nil, err
	}

	return &Metrics{
		&HTTP{requests: requests, duration: requestDuration},
		&Service{incomingMsgMetric, processedMetric, problemsSavingMetric},
	}, nil
}

// startHTTP запускает http сервер для связи с Prometheus на переданном в функцию порту и url. При неудаче выводит
// ошибку в лог и останавливает программу.
func startHTTP(url, port string) {
	go func() {
		mux := http.NewServeMux()
		mux.Handle(url, promhttp.Handler())
		slog.Info(fmt.Sprintf(":%s%s ready for prometheus", port, url))
		err := http.ListenAndServe(":"+port, mux)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("can't start http server for prometheus")
			os.Exit(1)
		}
	}()
}
