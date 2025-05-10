package application

import (
	"encoding/json"

	"github.com/anatoly_dev/go-ws-notifications/internal/domain"
	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
)

type KafkaHandler struct {
	notificationService domain.NotificationService
	logger              *logger.Logger
}

func NewKafkaHandler(notificationService domain.NotificationService, logger *logger.Logger) *KafkaHandler {
	return &KafkaHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

func (h *KafkaHandler) HandleMessage(message []byte) error {
	ctx := h.logger.WithField("source", "kafka_handler")

	ctx.Info("Получено новое сообщение из Kafka")

	var notification domain.Notification
	if err := json.Unmarshal(message, &notification); err != nil {
		ctx.WithError(err).Error("Ошибка десериализации сообщения из Kafka")
		return err
	}

	ctx = ctx.WithFields(map[string]interface{}{
		"notificationID": notification.ID,
		"userID":         notification.UserID,
		"type":           notification.Type,
	})

	ctx.Info("Обработка уведомления из Kafka")

	if err := h.notificationService.Send(&notification); err != nil {
		ctx.WithError(err).Error("Ошибка отправки уведомления")
		return err
	}

	ctx.Info("Уведомление из Kafka успешно обработано")
	return nil
}
