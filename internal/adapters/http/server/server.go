package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/service"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
)

// Server структура для обработки http-запросов к приложению.
type Server struct {
	cfg     *config.HttpServer // Конфигурация http сервера
	srv     *http.Server       // Структура с параметрами сервера
	mux     *http.ServeMux     // Мультиплексор http запросов
	service *service.Service   // Структура, реализующая логику приложения
}

// MustCreate создает и возвращает http-сервер.
func MustCreate(domainService *service.Service, cfg config.HttpServer) *Server {
	mux := http.NewServeMux()
	server := &Server{mux: mux, service: domainService, cfg: &cfg}

	mux.HandleFunc("/msg", server.ProcessMessage)
	// TODO использовать метрики Prometheus для сбора статистики?
	mux.HandleFunc("/statistic", server.Statistic)

	server.srv = &http.Server{
		Addr:         server.cfg.Address,
		Handler:      server.mux,
		ReadTimeout:  server.cfg.ReadTimeout,
		WriteTimeout: server.cfg.WriteTimeout,
		IdleTimeout:  server.cfg.IdleTimeout,
	}

	return server
}

// ProcessMessage ручка сохранения и отправки сообщения в брокер. Сообщение - содержимое тела запроса.
func (s *Server) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	if !allowedOnlyMethod(http.MethodPost, w, r) {
		return
	}

	var (
		message []byte
		err     error
		id      uuid.UUID
	)

	if message, err = ioutil.ReadAll(r.Body); err != nil || len(message) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err = s.service.ProcessMessage(r.Context(), message)
	if err == service.ErrSavingToRepository {
		w.WriteHeader(http.StatusProcessing)
		_, err = w.Write([]byte(fmt.Sprintf("{\"status\":\"temporaly problem to save\",\"msg_id\":%s}", id)))

		if err != nil {
			slog.Error(err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusProcessing)
	_, err = w.Write([]byte(fmt.Sprintf("{\"status\":\"saved, sent to the broker...\",\"msg_id\":%s}", id)))
	if err != nil {
		slog.Error(err.Error())
	}

}

// Statistic возвращает статистику пришедших/отправленных на временное хранение сообщений.
func (s *Server) Statistic(w http.ResponseWriter, r *http.Request) {
	if !allowedOnlyMethod(http.MethodGet, w, r) {
		return
	}

	statistic := s.service.Statistic()
	jsonData, err := json.Marshal(statistic)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		slog.Error(err.Error())
	}
}

// MustRun производит запуск сервера в отдельной go-рутине. В случае ошибки останавливает работу приложения.
func (s *Server) MustRun() {
	go func() {
		slog.Info("start http server on " + s.srv.Addr)
		err := s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server err: startup error. Initial error: " + err.Error())
			os.Exit(1)
		}
	}()
}

// Shutdown производит остановку сервера.
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		slog.Error("failed to gracefully shutdown http server")
	} else {
		slog.Info("gracefully shut down http server")
	}
}

// allowedOnlyMethod принимает разрешенный метод и, если запрос ему не соответствует, записывает в заголовок информацию
// о разрешенном методе, статус http.StatusMethodNotAllowed и возвращает false.
func allowedOnlyMethod(method string, w http.ResponseWriter, r *http.Request) bool {
	if r.Method != method {
		w.Header().Set("Allow", method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		slog.Default().With("remote address", r.RemoteAddr).With("request url", r.RequestURI).Warn("method not allowed")
		return false
	}

	return true
}
