/*
Package id_outbox: наивная реализация интерфейса "github.com/lazylex/messaggio/internal/ports/record_outbox"
*/

package record_outbox

import (
	"github.com/lazylex/messaggio/internal/dto"
	"sync"
)

const initSize = 10

type Naive struct {
	mu   sync.Mutex
	data []dto.MessageID
}

// New возвращает структуру для работы с outbox'ом. Для идентификации в логах используется name.
func New() *Naive {
	return &Naive{data: make([]dto.MessageID, 0, initSize)}
}

// Add добавление записи в outbox.
func (n *Naive) Add(data dto.MessageID) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.data = append(n.data, data)

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

	return result
}

// IsEmpty возвращает true, если outbox пуст.
func (n *Naive) IsEmpty() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.data) == 0
}
