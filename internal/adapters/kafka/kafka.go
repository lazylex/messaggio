package kafka

import (
	"github.com/lazylex/messaggio/internal/adapters/kafka/consumers/status"
	"github.com/lazylex/messaggio/internal/adapters/kafka/producers/message"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/ports/service"
	"log/slog"
	"os"
)

// MustRun запускает опрос/запись в топики Кафки.
func MustRun(cfg config.Kafka, service service.Interface, instance string) {
	if len(cfg.Brokers) == 0 {
		LogFatal("kafka broker list is empty")
	}
	if len(cfg.MessageTopic) == 0 {
		LogFatal("kafka status topic name is empty")
	}
	if len(cfg.ConfirmTopic) == 0 {
		LogFatal("kafka confirm topic name is empty")
	}

	status.StartInteraction(cfg, service, instance)
	message.StartInteraction(cfg, service, instance)
}

func LogFatal(reason string) {
	slog.Error(reason)
	os.Exit(1)
}
