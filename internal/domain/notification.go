package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type NotificationType string

const (
	TypeMessage NotificationType = "message"
	TypeSystem  NotificationType = "system"
	TypeAlert   NotificationType = "alert"
)

type Notification struct {
	ID        string           `json:"id" validate:"required"`
	UserID    string           `json:"user_id" validate:"required"`
	Type      NotificationType `json:"type" validate:"required,oneof=system alert message"`
	Title     string           `json:"title" validate:"required"`
	Content   string           `json:"content" validate:"required"`
	IsRead    bool             `json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
	Priority  int              `json:"priority" validate:"min=0,max=5"`
}

func (n *Notification) Validate() error {
	validate := validator.New()
	return validate.Struct(n)
}

type NotificationService interface {
	Send(notification *Notification) error
	MarkAsRead(id string, userID string) error
}

type NotificationRepository interface {
	Save(notification *Notification) error
	FindByID(id string) (*Notification, error)
	FindByUserID(userID string) ([]*Notification, error)
	Update(notification *Notification) error
}

type WebSocketService interface {
	SendToUser(userID string, message []byte) error
	BroadcastMessage(message []byte) error
}

type KafkaConsumer interface {
	Subscribe(topic string, handler func(message []byte) error) error
	Close() error
}
