package audit

import "github.com/Shopify/sarama"

// NewProducer create new audit producer
func NewProducer(brokers []string) (sarama.AsyncProducer, error) {
	sc := sarama.NewConfig()
	producer, err := sarama.NewAsyncProducer(brokers, sc)
	if err != nil {
		return nil, err
	}
	return producer, nil
}
