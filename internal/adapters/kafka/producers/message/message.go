package message

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/dto"
	"github.com/lazylex/messaggio/internal/ports/service"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"time"
)

var ErrMarshalJson = errors.New("failed to marshal message")

func StartInteraction(cfg config.Kafka, s service.Interface, instance string) {
	var err error
	var msg []byte

	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.MessageTopic,
		AllowAutoTopicCreation: true,
	}

	ch := s.MessageChan()
	go func() {
		defer func() {
			if err = w.Close(); err != nil {
				slog.Error(err.Error())
			}
		}()

		for {
			msgData := <-ch

			data := dto.MessageIdInstance{
				Message:  msgData.Message,
				ID:       msgData.ID,
				Instance: instance,
			}

			ctx, cancel := context.WithTimeout(context.Background(), cfg.KafkaWriteTimeout)

			if msg, err = json.Marshal(data); err != nil {
				slog.Error(ErrMarshalJson.Error())
				cancel()
				continue
			}

			err = w.WriteMessages(ctx, kafka.Message{Value: msg})
			if err == nil {
				cancel()
				continue
			}

			slog.Error(err.Error())

			if err = s.SaveUnsentMessage(msgData); err != nil {
				slog.Error(err.Error())
			}

			time.Sleep(cfg.KafkaTimeBetweenAttempts)
			cancel()
		}
	}()

}
