package websocket

import (
	"sync"

	"github.com/anatoly_dev/go-ws-notifications/internal/domain"
	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
)

type Service struct {
	clients     map[string]*Client
	clientsLock sync.RWMutex
	logger      *logger.Logger
	config      *Config
}

type Config struct {
	ReadBufferSize  int
	WriteBufferSize int
	PongWait        int
	PingPeriod      int
	MaxMessageSize  int64
}

func NewService(config *Config, logger *logger.Logger) *Service {
	return &Service{
		clients: make(map[string]*Client),
		logger:  logger,
		config:  config,
	}
}

func (s *Service) RegisterClient(userID string, client *Client) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	if existingClient, ok := s.clients[userID]; ok {
		s.logger.WithField("userID", userID).Info("Закрытие существующего соединения для пользователя")
		existingClient.Close()
	}

	s.clients[userID] = client
	s.logger.WithField("userID", userID).Info("Пользователь подключен к WebSocket")
}

func (s *Service) UnregisterClient(userID string) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	if _, ok := s.clients[userID]; ok {
		delete(s.clients, userID)
		s.logger.WithField("userID", userID).Info("Пользователь отключен от WebSocket")
	}
}

func (s *Service) SendToUser(userID string, message []byte) error {
	s.clientsLock.RLock()
	client, ok := s.clients[userID]
	s.clientsLock.RUnlock()

	if !ok {
		s.logger.WithField("userID", userID).Warn("Попытка отправки сообщения отключенному пользователю")
		return domain.ErrUserNotConnected
	}

	return client.Send(message)
}

func (s *Service) BroadcastMessage(message []byte) error {
	s.clientsLock.RLock()
	defer s.clientsLock.RUnlock()

	s.logger.WithField("clients_count", len(s.clients)).Info("Отправка широковещательного сообщения")

	for userID, client := range s.clients {
		if err := client.Send(message); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"userID": userID,
				"error":  err.Error(),
			}).Error("Ошибка отправки сообщения клиенту")
		}
	}

	return nil
}

func (s *Service) GetClientCount() int {
	s.clientsLock.RLock()
	defer s.clientsLock.RUnlock()
	return len(s.clients)
}
