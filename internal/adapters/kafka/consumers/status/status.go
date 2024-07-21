package status

import (
	"context"
	"encoding/json"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/dto"
	"github.com/lazylex/messaggio/internal/ports/service"
	"github.com/segmentio/kafka-go"
	"log/slog"
)

// StartInteraction запускает чтение и обработку сообщений из topic. Если instance в сообщении из топика не
// соответствует переданному в параметре функции, дальнейшая обработка сообщения не производится.
func StartInteraction(cfg config.Kafka, service service.Interface, instance string) {
	var err error
	var m kafka.Message

	ctx := context.Background()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.ConfirmTopic,
		MaxBytes: 10e6,
		GroupID:  instance,
	})

	go func() {
		defer func(r *kafka.Reader) {
			if err = r.Close(); err != nil {
				slog.Warn(err.Error())
			}
		}(r)

		for {
			if m, err = r.FetchMessage(ctx); err != nil {
				continue
			}

			var data dto.InstanceId

			if err = json.Unmarshal(m.Value, &data); err != nil {
				slog.Warn("error unmarshal JSON")
				continue
			}

			if data.Instance != instance {
				continue
			}

			if err = service.MarkMessageAsProcessed(ctx, data.ID); err != nil {
				slog.Warn(err.Error())
			} else if err = r.CommitMessages(ctx, m); err != nil {
				slog.Warn(err.Error())
			}

			// TODO учитывая, что не сохраненные в БД статусы не commit'ятся, стоит убрать outbox статусов. Он бесполезен
		}
	}()
}
