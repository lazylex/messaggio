package dto

type Statistic struct {
	Total                      uint64 `json:"total"`                         // Всего пришло сообщений на обработку
	MessagesSentToOutbox       uint64 `json:"messages_sent_to_outbox"`       // Всего сохранено сообщений в outbox
	MessagesReturnedFromOutbox uint64 `json:"messages_returned_from_outbox"` // Всего удалось переместить сообщений из outbox в БД
}
