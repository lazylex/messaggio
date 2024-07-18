package repo_outbox

import "github.com/lazylex/messaggio/internal/dto"

//go:generate mockgen -source=repo_outbox.go -destination=mocks/repo_outbox.go
type Interface interface {
	Add(id dto.MessageID) error
	Pop() dto.MessageID
}
