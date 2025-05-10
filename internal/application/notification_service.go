package application

import (
	"encoding/json"
	"time"

	"github.com/anatoly_dev/go-ws-notifications/internal/domain"
	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
	"github.com/google/uuid"
)

type NotificationService struct {
	repository domain.NotificationRepository
	wsService  domain.WebSocketService
	logger     *logger.Logger
}

func NewNotificationService(
	repository domain.NotificationRepository,
	wsService domain.WebSocketService,
	logger *logger.Logger,
) *NotificationService {
	return &NotificationService{
		repository: repository,
		wsService:  wsService,
		logger:     logger,
	}
}

func (s *NotificationService) Send(notification *domain.Notification) error {
	ctx := s.logger.WithFields(map[string]interface{}{
		"notificationID": notification.ID,
		"userID":         notification.UserID,
		"type":           notification.Type,
	})

	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	err := notification.Validate()
	if err != nil {
		ctx.WithError(err).Error("Ошибка валидации уведомления")
		return err
	}

	err = s.repository.Save(notification)
	if err != nil {
		ctx.WithError(err).Error("Ошибка сохранения уведомления")
		return err
	}

	message, err := json.Marshal(notification)
	if err != nil {
		ctx.WithError(err).Error("Ошибка сериализации уведомления")
		return err
	}

	err = s.wsService.SendToUser(notification.UserID, message)
	if err != nil {
		ctx.WithError(err).Error("Ошибка отправки уведомления через WebSocket")
		return err
	}

	ctx.Info("Уведомление успешно отправлено")
	return nil
}

func (s *NotificationService) MarkAsRead(id string, userID string) error {
	ctx := s.logger.WithFields(map[string]interface{}{
		"notificationID": id,
		"userID":         userID,
	})

	notification, err := s.repository.FindByID(id)
	if err != nil {
		ctx.WithError(err).Error("Ошибка поиска уведомления")
		return err
	}

	if notification.UserID != userID {
		ctx.Error("Попытка отметить чужое уведомление как прочитанное")
		return domain.ErrUnauthorized
	}

	notification.IsRead = true
	err = s.repository.Update(notification)
	if err != nil {
		ctx.WithError(err).Error("Ошибка обновления статуса уведомления")
		return err
	}

	ctx.Info("Уведомление отмечено как прочитанное")
	return nil
}
