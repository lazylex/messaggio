/*
Package id_outbox: наивная реализация интерфейса "github.com/lazylex/messaggio/internal/ports/record_outbox"
*/

package record_outbox

import (
	"fmt"
	"github.com/lazylex/messaggio/internal/dto"
	"log/slog"
	"sync"
)

const initSize = 10

type Naive struct {
	mu   sync.Mutex
	data []dto.MessageID
	name string
}

// New возвращает структуру для работы с outbox'ом. Для идентификации в логах используется name.
func New(name string) *Naive {
	return &Naive{name: name, data: make([]dto.MessageID, 0, initSize)}
}

// Add добавление записи в outbox.
func (n *Naive) Add(data dto.MessageID) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.data = append(n.data, data)

	slog.Info(
		fmt.Sprintf("Added to outbox: %s, %s", data.ID, data.Message),
		slog.String("name", n.name),
	)

	return nil
}

// Pop извлечение записи из outbox'а.
func (n *Naive) Pop() dto.MessageID {
	result := dto.MessageID{}
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.data) == 0 {
		return result
	}

	result, n.data = n.data[len(n.data)-1], n.data[:len(n.data)-1]

	slog.Info(
		fmt.Sprintf("Pop from outbox: %s, %s. Remaining length: %d", result.ID, result.Message, len(n.data)),
		slog.String("name", n.name),
	)

	return result
}

// Len возвращает количество элементов в хранилище.
func (n *Naive) Len() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.data)
}
