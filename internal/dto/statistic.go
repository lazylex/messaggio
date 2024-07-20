package dto

type Statistic struct {
	Total                uint64 `json:"total"`                   // Всего пришло сообщений на обработку
	StatusesSentToOutbox uint64 `json:"statuses_sent_to_outbox"` // Всего сохранено статусов в outbox
	MessagesSentToOutbox uint64 `json:"messages_sent_to_outbox"` // Всего сохранено сообщений в outbox
}
