package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
	"github.com/lazylex/messaggio/internal/dto"
	ido "github.com/lazylex/messaggio/internal/ports/id_outbox"
	reo "github.com/lazylex/messaggio/internal/ports/record_outbox"
	"github.com/lazylex/messaggio/internal/ports/repository"
	"log/slog"
	"os"
)

var (
	ErrSavingToRepository       = errors.New("failed to save to repository")
	ErrUpdateStatusInRepository = errors.New("failed to update status in repository")
	ErrSavingToRepoRecordOutbox = errors.New("failed to save to repository record outbox")
	ErrSavingToRepoStatusOutbox = errors.New("failed to save to repository status outbox")
)

type Service struct {
	repo        repository.Interface // Объект для взаимодействия с БД
	messageChan chan dto.MessageID   // Канал для отправки сообщений
	errorChan   chan uuid.UUID       // Канал для приема идентификаторов неотправленных сообщений
	outbox                           // Хранилище неотправленных данных
}

type outbox struct {
	broker     ido.Interface // Outbox для сохранения ID сообщений, не отправленных в брокер сообщений
	repoRecord reo.Interface // Outbox для сохранения сообщений с ID, не сохраненных в БД
	repoStatus ido.Interface // Outbox для сохранения ID сообщений, у которых не удалось обновить статус в БД
}

// MustCreate возвращает структуры для работы с сервисной логикой.
func MustCreate(repo repository.Interface, brokerOutbox, statusOutbox ido.Interface, repoOutbox reo.Interface) *Service {
	if repo == nil || brokerOutbox == nil || repoOutbox == nil || statusOutbox == nil {
		slog.Error("nil pointer in function parameters")
		os.Exit(1)
	}

	messageChan := make(chan dto.MessageID)
	errorChan := make(chan uuid.UUID)

	return &Service{messageChan: messageChan, errorChan: errorChan,
		outbox: outbox{broker: brokerOutbox, repoRecord: repoOutbox, repoStatus: statusOutbox},
		repo:   repo}
}

// ProcessMessage сохраняет сообщение в БД, затем отправляет его в Kafka. При ошибке сохранения в БД или отправки
// сообщения, оно сохраняется для последующих попыток записи в БД/отправки сообщения.
func (s *Service) ProcessMessage(ctx context.Context, msg message.Message) (uuid.UUID, error) {
	var err error

	id := uuid.New()
	data := dto.MessageID{Message: msg, ID: id}

	if err = s.saveMessage(ctx, data); err != nil {
		defer func() {
			if err := s.outbox.repoRecord.Add(data); err != nil {
				slog.Error(err.Error())
			}
		}()

		return id, ErrSavingToRepository
	}

	defer func() {
		s.messageChan <- data
	}()

	return id, nil
}

// MarkMessageAsProcessed меняет статус в БД у сообщения на "Processed".
func (s *Service) MarkMessageAsProcessed(ctx context.Context, id uuid.UUID) error {
	var err error

	if err = s.repo.UpdateStatus(ctx, id); err == nil {
		return nil
	}

	if err = s.outbox.repoStatus.Add(id); err != nil {
		slog.Error(ErrSavingToRepoStatusOutbox.Error())
		return err
	}

	return ErrUpdateStatusInRepository
}

// saveMessage сохраняет сообщение в БД. При ошибке сохранения записывает в outbox для дальнейших попыток сохранения.
func (s *Service) saveMessage(ctx context.Context, data dto.MessageID) error {
	var err error

	if err = s.repo.SaveMessage(ctx, data); err == nil {
		return nil
	}

	if err = s.outbox.repoRecord.Add(data); err != nil {
		slog.Error(ErrSavingToRepoRecordOutbox.Error())
		return err
	}

	return ErrSavingToRepository
}

// MessageChan возвращает канал, который будет служить для отправки сообщений в брокер сообщений.
func (s *Service) MessageChan() chan dto.MessageID {
	return s.messageChan
}

// ErrorChan возвращает канал для приема идентификаторов сообщений, которые не удалось отправить в топик.
func (s *Service) ErrorChan() chan uuid.UUID {
	return s.errorChan
}
