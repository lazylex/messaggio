/*
Package id_outbox: наивная реализация интерфейса "github.com/lazylex/messaggio/internal/ports/id_outbox"
*/

package id_outbox

import (
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"sync"
)

const initSize = 10

type Naive struct {
	mu   sync.Mutex
	ids  []uuid.UUID
	name string
}

// New возвращает структуру для работы с outbox'ом. Для идентификации в логах используется name.
func New(name string) *Naive {
	return &Naive{name: name, ids: make([]uuid.UUID, 0, initSize)}
}

// Add добавление записи в outbox.
func (n *Naive) Add(uuid uuid.UUID) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ids = append(n.ids, uuid)

	slog.Info(
		fmt.Sprintf("Added to outbox: %s", uuid.String()),
		slog.String("name", n.name),
	)

	return nil
}

// Pop извлечение записи из outbox'а.
func (n *Naive) Pop() uuid.UUID {
	result := uuid.Nil
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.ids) == 0 {
		return result
	}

	result, n.ids = n.ids[len(n.ids)-1], n.ids[:len(n.ids)-1]

	slog.Info(
		fmt.Sprintf("Pop from outbox: %s. Remaining length: %d", result.String(), len(n.ids)),
		slog.String("name", n.name),
	)

	return result
}

// Len возвращает количество элементов в хранилище.
func (n *Naive) Len() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.ids)
}
