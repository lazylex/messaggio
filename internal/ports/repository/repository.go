package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/dto"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go
type Interface interface {
	SaveMessage(ctx context.Context, data dto.MessageID) error
	UpdateStatus(ctx context.Context, id uuid.UUID) error
}
