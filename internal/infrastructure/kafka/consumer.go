package kafka

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	logger  *logger.Logger
	handler func(message []byte) error
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	config  *ConsumerConfig
}

type ConsumerConfig struct {
	Brokers         []string
	GroupID         string
	AutoOffsetReset string
}

func NewConsumer(config *ConsumerConfig, logger *logger.Logger) (*Consumer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		config: config,
	}, nil
}

func (c *Consumer) Subscribe(topic string, handler func(message []byte) error) error {
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.config.Brokers,
		Topic:       topic,
		GroupID:     c.config.GroupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})

	c.handler = handler
	c.logger.WithField("topic", topic).Info("Подписка на топик Kafka успешно установлена")

	c.wg.Add(1)
	go c.consumeMessages()

	return nil
}

func (c *Consumer) consumeMessages() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Остановка потребителя Kafka")
			return
		default:
			message, err := c.reader.ReadMessage(c.ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				c.logger.WithError(err).Error("Ошибка чтения сообщения из Kafka")
				time.Sleep(time.Second)
				continue
			}

			c.logger.WithFields(map[string]interface{}{
				"topic":     message.Topic,
				"partition": message.Partition,
				"offset":    message.Offset,
			}).Debug("Получено сообщение из Kafka")

			if err := c.handler(message.Value); err != nil {
				c.logger.WithError(err).Error("Ошибка обработки сообщения из Kafka")
			}
		}
	}
}

func (c *Consumer) Close() error {
	c.cancel()
	if c.reader != nil {
		c.reader.Close()
	}
	c.wg.Wait()
	c.logger.Info("Соединение с Kafka закрыто")
	return nil
}
