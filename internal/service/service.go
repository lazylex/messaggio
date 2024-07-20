package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
	"github.com/lazylex/messaggio/internal/dto"
	ido "github.com/lazylex/messaggio/internal/ports/id_outbox"
	reo "github.com/lazylex/messaggio/internal/ports/record_outbox"
	"github.com/lazylex/messaggio/internal/ports/repository"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

var (
	ErrSavingToRepository       = errors.New("failed to save to repository")
	ErrUpdateStatusInRepository = errors.New("failed to update status in repository")
	ErrSavingToRepoRecordOutbox = errors.New("failed to save to repository record outbox")
	ErrSavingToRepoStatusOutbox = errors.New("failed to save to repository status outbox")
)

type Service struct {
	repo                       repository.Interface // Объект для взаимодействия с БД
	messageChan                chan dto.MessageID   // Канал для отправки сообщений
	errorChan                  chan uuid.UUID       // Канал для приема идентификаторов неотправленных сообщений
	outbox                     outbox               // Хранилище неотправленных данных
	total                      atomic.Uint64        // Всего пришло сообщений на обработку
	statusesSentToOutbox       atomic.Uint64        // Всего сохранено статусов в outbox
	statusesReturnedFromOutbox atomic.Uint64        // Всего удалось переместить данных о статусе из outbox в БД
	messagesSentToOutbox       atomic.Uint64        // Всего сохранено сообщений в outbox
	messagesReturnedFromOutbox atomic.Uint64        // Всего удалось переместить сообщений из outbox в БД
}

type outbox struct {
	repoRecord reo.Interface // Outbox для сохранения сообщений с ID, не сохраненных в БД
	repoStatus ido.Interface // Outbox для сохранения ID сообщений, у которых не удалось обновить статус в БД
}

// MustCreate возвращает структуры для работы с сервисной логикой.
func MustCreate(repo repository.Interface, statusOutbox ido.Interface, repoOutbox reo.Interface, cfg config.Service) *Service {
	if repo == nil || repoOutbox == nil || statusOutbox == nil {
		slog.Error("nil pointer in function parameters")
		os.Exit(1)
	}

	messageChan := make(chan dto.MessageID)
	errorChan := make(chan uuid.UUID)

	s := &Service{messageChan: messageChan, errorChan: errorChan,
		outbox: outbox{repoRecord: repoOutbox, repoStatus: statusOutbox},
		repo:   repo}

	go func() {
		for range time.Tick(cfg.RetryTimeout) {
			go s.trySaveMessageAgain()
			go s.tryUpdateMessageStatusAgain()
		}
	}()

	return s
}

// ProcessMessage сохраняет сообщение в БД, затем отправляет его в Kafka. При ошибке сохранения в БД или отправки
// сообщения, оно сохраняется для последующих попыток записи в БД/отправки сообщения.
func (s *Service) ProcessMessage(ctx context.Context, msg message.Message) (uuid.UUID, error) {
	var err error

	id := uuid.New()
	data := dto.MessageID{Message: msg, ID: id}

	s.total.Add(1)

	if err = s.saveMessage(ctx, data); err != nil {
		defer func() {
			if err := s.outbox.repoRecord.Add(data); err != nil {
				slog.Error(err.Error())
				return
			}

			s.messagesSentToOutbox.Add(1)
		}()

		return id, ErrSavingToRepository
	}

	go func() {
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

	s.statusesSentToOutbox.Add(1)

	return ErrUpdateStatusInRepository
}

// trySaveMessageAgain рекурсивно пытается сохранить в БД сообщения, ранее сохраненные в outbox. Попытки осуществляются
// пока outbox содержит элементы и сохранение не вызывает ошибку.
func (s *Service) trySaveMessageAgain() {
	var err error
	ctx := context.Background()

	if s.outbox.repoRecord.Len() == 0 {
		return
	}

	record := s.outbox.repoRecord.Pop()
	if err = s.repo.SaveMessage(ctx, record); err == nil {
		s.messagesReturnedFromOutbox.Add(1)
		s.trySaveMessageAgain()
	} else {
		err = s.outbox.repoRecord.Add(record)
		if err != nil {
			slog.Error(err.Error())
		}
	}
}

// tryUpdateMessageStatusAgain рекурсивно пытается обновить в БД статус сообщений, чьи идентификаторы ранее были
// сохранены в outbox. Попытки осуществляются пока outbox содержит элементы и обновление не вызывает ошибку.
func (s *Service) tryUpdateMessageStatusAgain() {
	var err error
	ctx := context.Background()

	if s.outbox.repoStatus.Len() == 0 {
		return
	}

	id := s.outbox.repoStatus.Pop()
	if err = s.repo.UpdateStatus(ctx, id); err == nil {
		s.statusesReturnedFromOutbox.Add(1)
		s.tryUpdateMessageStatusAgain()
	} else {
		err = s.outbox.repoStatus.Add(id)
		if err != nil {
			slog.Error(err.Error())
		}
	}
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

// Statistic возвращает статистику.
func (s *Service) Statistic() dto.Statistic {
	return dto.Statistic{
		Total:                      s.total.Load(),
		StatusesSentToOutbox:       s.statusesSentToOutbox.Load(),
		MessagesSentToOutbox:       s.messagesSentToOutbox.Load(),
		MessagesReturnedFromOutbox: s.messagesReturnedFromOutbox.Load(),
		StatusesReturnedFromOutbox: s.statusesReturnedFromOutbox.Load(),
	}
}

// MessageChan возвращает канал, который будет служить для отправки сообщений в брокер сообщений.
func (s *Service) MessageChan() chan dto.MessageID {
	return s.messageChan
}

// ErrorChan возвращает канал для приема идентификаторов сообщений, которые не удалось отправить в топик.
func (s *Service) ErrorChan() chan uuid.UUID {
	return s.errorChan
}
