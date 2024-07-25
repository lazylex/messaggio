package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
	"github.com/lazylex/messaggio/internal/dto"
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
