package kafka

import (
	"time"

	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
	logger   *logger.Logger
}

type ConsumerConfig struct {
	Brokers         []string
	GroupID         string
	AutoOffsetReset string
}

func NewConsumer(config *ConsumerConfig, logger *logger.Logger) (*Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  config.Brokers[0],
		"group.id":           config.GroupID,
		"auto.offset.reset":  config.AutoOffsetReset,
		"enable.auto.commit": true,
	})

	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		logger:   logger,
	}, nil
}

func (c *Consumer) Subscribe(topic string, handler func(message []byte) error) error {
	err := c.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		c.logger.WithError(err).Error("Ошибка подписки на топик Kafka")
		return err
	}

	c.logger.WithField("topic", topic).Info("Подписка на топик Kafka успешно установлена")

	go c.consumeMessages(handler)

	return nil
}

func (c *Consumer) consumeMessages(handler func(message []byte) error) {
	for {
		msg, err := c.consumer.ReadMessage(time.Second * 1)
		if err != nil {

			if err.(kafka.Error).Code() == kafka.ErrTimedOut {
				continue
			}

			c.logger.WithError(err).Error("Ошибка чтения сообщения из Kafka")
			continue
		}

		c.logger.WithFields(map[string]interface{}{
			"topic":     *msg.TopicPartition.Topic,
			"partition": msg.TopicPartition.Partition,
			"offset":    msg.TopicPartition.Offset,
		}).Debug("Получено сообщение из Kafka")

		if err := handler(msg.Value); err != nil {
			c.logger.WithError(err).Error("Ошибка обработки сообщения из Kafka")
		}
	}
}

func (c *Consumer) Close() error {
	c.logger.Info("Закрытие соединения с Kafka")
	return c.consumer.Close()
}
