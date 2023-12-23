package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

func NewConsumerGroup(ctx context.Context, consumerGroupId string, addresses []string) (sarama.ConsumerGroup, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.DefaultVersion
	kafkaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(addresses, consumerGroupId, kafkaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new consumer group sarama error")
	}

	return consumerGroup, nil
}
