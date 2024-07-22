package metrics

import "github.com/prometheus/client_golang/prometheus"

type Service struct {
	incomingMsgInc     *prometheus.CounterVec
	processedMsgInc    *prometheus.CounterVec
	problemsSavingInDB *prometheus.CounterVec
}

// IncomingMsgInc увеличивает счетчик пришедших по HTTP сообщений.
func (s *Service) IncomingMsgInc() {
	s.incomingMsgInc.With(prometheus.Labels{}).Inc()
}

// ProcessedMsgInc увеличивает счетчик успешно обработанных сообщений.
func (s *Service) ProcessedMsgInc() {
	s.processedMsgInc.With(prometheus.Labels{}).Inc()
}

// ProblemsSavingInDB счетчик неудачных сохранений в БД.
func (s *Service) ProblemsSavingInDB() {
	s.problemsSavingInDB.With(prometheus.Labels{}).Inc()
}

// createIncomingMsgTotalMetric создает и регистрирует метрику incoming_messages_total, являющуюся счетчиком пришедших
// в обработку сообщений.
func createIncomingMsgTotalMetric() (*prometheus.CounterVec, error) {
	var err error
	orders := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "incoming_messages_total",
		Namespace: NAMESPACE,
		Help:      "Count of incoming messages",
	}, []string{})
	if err = prometheus.Register(orders); err != nil {
		return nil, err
	}

	orders.With(prometheus.Labels{})

	return orders, nil
}

// createProcessedMsgTotalMetric создает и регистрирует метрику processed_messages_total, являющуюся счетчиком
// обработанных сообщений.
func createProcessedMsgTotalMetric() (*prometheus.CounterVec, error) {
	var err error
	orders := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "processed_messages_total",
		Namespace: NAMESPACE,
		Help:      "Count of processed messages",
	}, []string{})
	if err = prometheus.Register(orders); err != nil {
		return nil, err
	}

	orders.With(prometheus.Labels{})

	return orders, nil
}

// createProblemsSavingInDBTotalMetric создает и регистрирует метрику problems_saving_total, являющуюся счетчиком
// проблем при сохранении пришедших на обработку сообщений.
func createProblemsSavingInDBTotalMetric() (*prometheus.CounterVec, error) {
	var err error
	orders := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "problems_saving_total",
		Namespace: NAMESPACE,
		Help:      "Count of saving problems",
	}, []string{})
	if err = prometheus.Register(orders); err != nil {
		return nil, err
	}

	orders.With(prometheus.Labels{})

	return orders, nil
}
