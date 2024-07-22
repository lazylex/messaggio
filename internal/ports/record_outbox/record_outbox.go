package record_outbox

import "github.com/lazylex/messaggio/internal/dto"

//go:generate mockgen -source=record_outbox.go -destination=mocks/record_outbox.go
type Interface interface {
	Add(dto.MessageID) error
	Pop() dto.MessageID
	Len() int
	// TODO заменить Len на IsEmpty, так как для алгоритма количество элементов не имеет значения
}
