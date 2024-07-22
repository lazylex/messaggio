package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	srvc "github.com/lazylex/messaggio/internal/ports/service"
	"github.com/lazylex/messaggio/internal/service"
	"io/ioutil"
	"log/slog"
	"net/http"
	"time"
)

// Handler структура для обработки http-запросов.
type Handler struct {
	service      srvc.Interface // Объект, реализующий логику сервиса
	queryTimeout time.Duration  // Допустимый таймаут для обработки запроса
}

// New возвращает структуру с обработчиками http-запросов.
func New(domainService srvc.Interface, timeout time.Duration) *Handler {
	return &Handler{service: domainService, queryTimeout: timeout}
}

// ProcessMessage ручка сохранения и отправки сообщения в Kafka. Сообщение - содержимое тела запроса.
func (h *Handler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
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

	id, err = h.service.ProcessMessage(r.Context(), message)
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
func (h *Handler) Statistic(w http.ResponseWriter, r *http.Request) {
	if !allowedOnlyMethod(http.MethodGet, w, r) {
		return
	}

	statistic := h.service.Statistic()
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
