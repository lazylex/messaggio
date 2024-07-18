package broker_outbox

import "github.com/google/uuid"

//go:generate mockgen -source=broker_outbox.go -destination=mocks/broker_outbox.go
type Interface interface {
	Add(uuid uuid.UUID) error
	Pop() uuid.UUID
}
