package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Consumer struct {
	logEnable bool
	handler   func(ctx context.Context, message *sarama.ConsumerMessage) error
}

func NewConsumer(handler func(ctx context.Context, message *sarama.ConsumerMessage) error) Consumer {
	return Consumer{
		logEnable: false,
		handler:   handler,
	}
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}

			// message logging
			if consumer.logEnable {
				dst := &bytes.Buffer{}
				json.Compact(dst, message.Value)

				var headers = map[string]string{}
				for _, h := range message.Headers {
					headers[string(h.Key)] = string(h.Value)
				}

				log.Printf(
					"consume message: headers = %+v, value = %s, timestamp = %s, partition = %d, offset = %d\n",
					headers,
					dst.String(),
					message.Timestamp.Format(time.RFC3339),
					message.Partition,
					message.Offset,
				)
			}

			// start message processing
			err := consumer.handler(context.Background(), message)
			if err != nil {
				log.Printf("consumer handler error: %s", err.Error())
				return err
			}

			// mark message
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
