package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
	"github.com/lazylex/messaggio/internal/dto"
	"github.com/lazylex/messaggio/internal/ports/metrics/service"
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
)

type Service struct {
	repo                       repository.Interface     // Объект для взаимодействия с БД
	messageChan                chan dto.MessageID       // Канал для отправки сообщений
	outbox                     outbox                   // Хранилище неотправленных данных
	total                      atomic.Uint64            // Всего пришло сообщений на обработку
	messagesSentToOutbox       atomic.Uint64            // Всего сохранено сообщений в outbox
	messagesReturnedFromOutbox atomic.Uint64            // Всего удалось переместить сообщений из outbox в БД
	canRetrySendToBroker       atomic.Bool              // Флаг, означающий, что запись в канал для отправки сообщений в kafka прошла успешно и есть смысл запускать go-рутину для последующих попыток отправки
	metrics                    service.MetricsInterface // Метрики Prometheus
}

type outbox struct {
	brokerRecord reo.Interface // Outbox для сохранения сообщений с ID, не отправленных в Kafka
	repoRecord   reo.Interface // Outbox для сохранения сообщений с ID, не сохраненных в БД
}

// MustCreate возвращает структуры для работы с сервисной логикой.
func MustCreate(repo repository.Interface, brokerOutbox, repoOutbox reo.Interface,
	cfg config.Service, metrics service.MetricsInterface) *Service {
	if repo == nil || repoOutbox == nil || metrics == nil {
		slog.Error("nil pointer in function parameters")
		os.Exit(1)
	}

	messageChan := make(chan dto.MessageID)

	s := &Service{messageChan: messageChan,
		outbox:  outbox{repoRecord: repoOutbox, brokerRecord: brokerOutbox},
		repo:    repo,
		metrics: metrics,
	}

	s.canRetrySendToBroker.Store(true)
	go func() {
		for range time.Tick(cfg.RetryTimeout) {
			go s.trySaveMessageAgain()

			if s.canRetrySendToBroker.Load() {
				go s.trySendToBrokerAgain()
			}
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

	s.metrics.IncomingMsgInc()
	s.total.Add(1)

	if err = s.saveMessage(ctx, data); err != nil {
		defer func() {
			s.metrics.ProblemsSavingInDB()
			if err := s.outbox.repoRecord.Add(data); err != nil {
				slog.Error(err.Error())
				return
			}

			s.messagesSentToOutbox.Add(1)
		}()

		return id, ErrSavingToRepository
	}

	go func() {
		if s.outbox.brokerRecord.Len() > 0 {
			if err = s.outbox.brokerRecord.Add(data); err != nil {
				slog.Error(err.Error())
			}
			return
		}

		s.messageChan <- data
	}()

	return id, nil
}

// MarkMessageAsProcessed меняет статус в БД у сообщения на "Processed".
func (s *Service) MarkMessageAsProcessed(ctx context.Context, id uuid.UUID) error {
	if s.repo.UpdateStatus(ctx, id) == nil {
		s.metrics.ProcessedMsgInc()
		return nil
	}

	return ErrUpdateStatusInRepository
}

// SaveUnsentMessage сохраняет в outbox сообщение, которое не удалось отправить в брокер сообщений.
func (s *Service) SaveUnsentMessage(data dto.MessageID) error {
	return s.outbox.brokerRecord.Add(data)
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

// trySendToBrokerAgain рекурсивно пытается отправить в брокер не отправленное ранее сообщение. Попытки осуществляются
// пока outbox содержит элементы.
func (s *Service) trySendToBrokerAgain() {
	if s.outbox.brokerRecord.Len() == 0 {
		return
	}

	data := s.outbox.brokerRecord.Pop()
	s.canRetrySendToBroker.Store(false)
	s.messageChan <- data
	s.canRetrySendToBroker.Store(true)

	s.trySendToBrokerAgain()
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
		MessagesSentToOutbox:       s.messagesSentToOutbox.Load(),
		MessagesReturnedFromOutbox: s.messagesReturnedFromOutbox.Load(),
	}
}

// MessageChan возвращает канал, который будет служить для отправки сообщений в брокер сообщений.
func (s *Service) MessageChan() chan dto.MessageID {
	return s.messageChan
}
