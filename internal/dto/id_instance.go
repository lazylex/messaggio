package dto

import (
	"github.com/google/uuid"
)

type InstanceId struct {
	ID       uuid.UUID `json:"id"`
	Instance string    `json:"instance"`
}
