package id_outbox

import "github.com/google/uuid"

//go:generate mockgen -source=id_outbox.go -destination=mocks/id_outbox.go
type Interface interface {
	Add(uuid uuid.UUID) error
	Pop() uuid.UUID
}
