package kafka

import (
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type Producer struct {
	SyncProducer sarama.SyncProducer
	logEnable    bool
}

func NewProducer(addresses []string) (Producer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.DefaultVersion
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 3
	kafkaConfig.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(addresses, kafkaConfig)
	if err != nil {
		return Producer{}, errors.Wrap(err, "new producer sarama error")
	}

	return Producer{
		SyncProducer: syncProducer,
	}, nil
}

func (producer Producer) SendMessage(topic string, message string, headers map[string]string) error {
	if headers == nil {
		headers = map[string]string{}
	}

	recordHeaders := []sarama.RecordHeader{}
	for keyHeader, valueHeader := range headers {
		recordHeaders = append(recordHeaders, sarama.RecordHeader{
			Key:   []byte(keyHeader),
			Value: []byte(valueHeader),
		})
	}

	producerMessage := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(message),
		Headers:   recordHeaders,
		Timestamp: time.Now(),
	}

	partition, offset, err := producer.SyncProducer.SendMessage(producerMessage)
	if err != nil {
		return errors.Wrap(err, "send message error")
	}

	// message logging
	if producer.logEnable {
		log.Printf("produce message: headers = %+v, value = %s, timestamp = %s, partition = %d, offset = %d\n",
			headers,
			message,
			producerMessage.Timestamp.Format(time.RFC3339),
			partition,
			offset,
		)
	}

	return nil
}

func (producer Producer) CloseConnection() error {
	return producer.SyncProducer.Close()
}
