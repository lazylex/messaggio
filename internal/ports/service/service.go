package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
	"github.com/lazylex/messaggio/internal/dto"
)

var (
	ErrSavingToRepository       = errors.New("service: failed to save to repository")
	ErrUpdateStatusInRepository = errors.New("service: failed to update status in repository")
	ErrSavingToRepoRecordOutbox = errors.New("service: failed to save to repository record outbox")
)

//go:generate mockgen -source=service.go -destination=mocks/service.go
type Interface interface {
	ProcessMessage(ctx context.Context, msg message.Message) (uuid.UUID, error)
	MarkMessageAsProcessed(ctx context.Context, id uuid.UUID) error
	MessageChan() chan dto.MessageID
	SaveUnsentMessage(dto.MessageID) error
	Statistic() dto.Statistic
	ProcessedCountStatistic(ctx context.Context) (dto.Processed, error)
}
