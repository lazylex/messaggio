package http

import (
	"github.com/gin-gonic/gin"
	srvc "github.com/lazylex/messaggio/internal/ports/service"
	"io/ioutil"
	"log/slog"
	"net/http"
)

// Handler структура для обработки http-запросов.
type Handler struct {
	service srvc.Interface // Объект, реализующий логику сервиса
}

// NewHandler возвращает структуру с обработчиками http-запросов.
func NewHandler(domainService srvc.Interface) *Handler {
	return &Handler{service: domainService}
}

// ProcessMessage ручка сохранения и отправки сообщения в Kafka. Сообщение - содержимое тела запроса.
func (h *Handler) ProcessMessage(c *gin.Context) {
	var message []byte
	var err error
	if message, err = ioutil.ReadAll(c.Request.Body); err != nil || len(message) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"problem": "can't read message from body"})
		return
	}

	id, errSave := h.service.ProcessMessage(c.Request.Context(), message)
	if errSave == srvc.ErrSavingToRepository {
		c.JSON(http.StatusProcessing, gin.H{"status": "temporally problem to save", "msg_id": id})
		return
	}

	c.JSON(http.StatusProcessing, gin.H{"status": "saved, sent to the broker...", "msg_id": id})
}

// Statistic возвращает статистику пришедших/отправленных на временное хранение сообщений.
func (h *Handler) Statistic(c *gin.Context) {
	statistic := h.service.Statistic()
	c.JSON(http.StatusOK, statistic)
}

// ProcessedStatistic возвращает статистику по обработанным сообщениям за час, день, неделю, месяц.
func (h *Handler) ProcessedStatistic(c *gin.Context) {
	statistic, err := h.service.ProcessedCountStatistic(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"problem": "can't get statistic"})
		slog.Error(err.Error())
		return
	}

	c.JSON(http.StatusOK, statistic)
}
