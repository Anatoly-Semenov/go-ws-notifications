package repository

import (
	"sync"

	"github.com/anatoly_dev/go-ws-notifications/internal/domain"
	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
)

type MemoryRepository struct {
	notifications map[string]*domain.Notification
	userIndex     map[string][]*domain.Notification
	mutex         sync.RWMutex
	logger        *logger.Logger
}

func NewMemoryRepository(logger *logger.Logger) *MemoryRepository {
	return &MemoryRepository{
		notifications: make(map[string]*domain.Notification),
		userIndex:     make(map[string][]*domain.Notification),
		logger:        logger,
	}
}

func (r *MemoryRepository) Save(notification *domain.Notification) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.logger.WithField("notificationID", notification.ID).Debug("Сохранение уведомления")

	r.notifications[notification.ID] = notification

	r.userIndex[notification.UserID] = append(r.userIndex[notification.UserID], notification)

	return nil
}

func (r *MemoryRepository) FindByID(id string) (*domain.Notification, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	notification, exists := r.notifications[id]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return notification, nil
}

func (r *MemoryRepository) FindByUserID(userID string) ([]*domain.Notification, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	notifications, exists := r.userIndex[userID]
	if !exists {
		return []*domain.Notification{}, nil
	}

	result := make([]*domain.Notification, len(notifications))
	copy(result, notifications)

	return result, nil
}

func (r *MemoryRepository) Update(notification *domain.Notification) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, exists := r.notifications[notification.ID]
	if !exists {
		return domain.ErrNotFound
	}

	r.notifications[notification.ID] = notification
	
	for i, n := range r.userIndex[notification.UserID] {
		if n.ID == notification.ID {
			r.userIndex[notification.UserID][i] = notification
			break
		}
	}

	r.logger.WithField("notificationID", notification.ID).Debug("Уведомление обновлено")

	return nil
}
