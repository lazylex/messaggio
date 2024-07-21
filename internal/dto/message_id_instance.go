package dto

import (
	"github.com/google/uuid"
	"github.com/lazylex/messaggio/internal/domain/value_objects/message"
)

type MessageIdInstance struct {
	Message  message.Message `json:"message"`
	ID       uuid.UUID       `json:"id"`
	Instance string          `json:"instance"`
}
