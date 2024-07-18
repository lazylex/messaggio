package repository

import (
	"context"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go
type Interface interface {
	SaveMessage(ctx context.Context, msg message.Message) error
	UpdateStatus(ctx context.Context)
}
