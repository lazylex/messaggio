package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go
type Interface interface {
	ProcessMessage(ctx context.Context, msg message.Message) (uuid.UUID, error)
}
