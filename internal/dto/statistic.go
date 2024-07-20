package dto

type Statistic struct {
	Total                      uint64 `json:"total"`                         // Всего пришло сообщений на обработку
	StatusesSentToOutbox       uint64 `json:"statuses_sent_to_outbox"`       // Всего сохранено статусов в outbox
	StatusesReturnedFromOutbox uint64 `json:"statuses_returned_from_outbox"` // Всего удалось переместить данных о статусе из outbox в БД
	MessagesSentToOutbox       uint64 `json:"messages_sent_to_outbox"`       // Всего сохранено сообщений в outbox
	MessagesReturnedFromOutbox uint64 `json:"messages_returned_from_outbox"` // Всего удалось переместить сообщений из outbox в БД
}
