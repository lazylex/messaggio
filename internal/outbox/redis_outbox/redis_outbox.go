package redis_outbox

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
	"github.com/lazylex/messaggio/internal/dto"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
)

const outboxPrefix = "rop"

type RedisOutbox struct {
	client   *redis.Client // Клиент redis-сервера
	instance string        // Уникальный идентификатор экземпляра приложения для генерации ключей
	name     string        // Уникальное имя экземпляра outbox'а
}

// MustCreate создание структуры с клиентом для взаимодействия с Redis. При ошибке соединения с сервером Redis выводит
// ошибку в лог и прекращает работу приложения.
func MustCreate(client *redis.Client, name, instance string) *RedisOutbox {

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	} else {
		slog.Info("successfully received pong from redis server")
	}

	return &RedisOutbox{client: client, instance: instance, name: name}
}

// Add добавляет сообщение и идентификатор в список.
func (ro *RedisOutbox) Add(data dto.MessageID) error {
	if len(string(data.Message)) == 0 || data.ID == uuid.Nil {
		return errors.New("data is empty")
	}

	ctx := context.Background()
	return ro.client.LPush(ctx, ro.key(), data.ID.String(), string(data.Message)).Err()
}

// Pop извлекает сообщение и идентификатор из списка. Если список пуст, возвращает пустую структуру.
func (ro *RedisOutbox) Pop() dto.MessageID {
	var id uuid.UUID
	ctx := context.Background()

	data, err := ro.client.LPopCount(ctx, ro.key(), 2).Result()

	if err != nil {
		return dto.MessageID{}
	}

	if id, err = uuid.Parse(data[1]); err != nil {
		return dto.MessageID{}
	}

	return dto.MessageID{ID: id, Message: message.Message(data[0])}
}

// Len возвращает длину списка.
func (ro *RedisOutbox) Len() int {
	result, err := ro.client.LLen(context.Background(), ro.key()).Result()
	if err != nil {
		return 0
	}

	return int(result)
}

// key возвращает ключ, по которому в Redis будут сохраняться данные в списке.
func (ro *RedisOutbox) key() string {
	return fmt.Sprintf("%s:%s:%s", outboxPrefix, ro.name, ro.instance)
}
