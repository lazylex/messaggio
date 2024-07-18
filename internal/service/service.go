package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
	"github.com/lazylex/messaggio/internal/dto"
	bo "github.com/lazylex/messaggio/internal/ports/broker_outbox"
	ro "github.com/lazylex/messaggio/internal/ports/repo_outbox"
	"github.com/lazylex/messaggio/internal/ports/repository"
	"log/slog"
)

type Service struct {
	repo        repository.Interface // Объект для взаимодействия с БД
	messageChan chan dto.MessageID   // Канал для отправки сообщенийM
	errorChan   chan uuid.UUID       // Канал для приема идентификаторов неотправленных сообщений
	outbox                           // Хранилище неотправленных данных
}

type outbox struct {
	broker     bo.Interface
	repository ro.Interface
}

// New возвращает структуры для работы с сервисной логикой.
func New(repo repository.Interface, brokerOutbox bo.Interface, repoOutbox ro.Interface) *Service {
	messageChan := make(chan dto.MessageID)
	errorChan := make(chan uuid.UUID)

	return &Service{messageChan: messageChan, errorChan: errorChan,
		outbox: outbox{broker: brokerOutbox, repository: repoOutbox},
		repo:   repo}
}

// ProcessMessage сохраняет сообщение в БД, затем отправляет его в Kafka. При ошибке сохранения в БД или отправки
// сообщения, оно сохраняется для последующих попыток записи в БД/отправки сообщения.
func (s *Service) ProcessMessage(ctx context.Context, msg message.Message) (uuid.UUID, error) {
	// TODO implement
	slog.Debug("service: ProcessMessage not implemented")
	return uuid.Nil, errors.New("not implemented")
}

// MessageChan возвращает канал, который будет служить для отправки сообщений в брокер сообщений.
func (s *Service) MessageChan() chan dto.MessageID {
	return s.messageChan
}

// ErrorChan возвращает канал для приема идентификаторов сообщений, которые не удалось отправить в топик.
func (s *Service) ErrorChan() chan uuid.UUID {
	return s.errorChan
}
