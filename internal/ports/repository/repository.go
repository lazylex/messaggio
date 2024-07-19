package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/dto"
)

var (
	ErrDuplicateKeyValue = errors.New("duplicate key value violates unique constraint violation")
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go
type Interface interface {
	SaveMessage(ctx context.Context, data dto.MessageID) error
	UpdateStatus(ctx context.Context, id uuid.UUID) error
}
