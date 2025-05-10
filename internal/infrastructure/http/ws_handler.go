package http

import (
	"net/http"

	"github.com/anatoly_dev/go-ws-notifications/internal/infrastructure/websocket"
	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
	gorillaWs "github.com/gorilla/websocket"
)

type WSHandler struct {
	wsService *websocket.Service
	upgrader  gorillaWs.Upgrader
	logger    *logger.Logger
	config    *websocket.Config
}

func NewWSHandler(wsService *websocket.Service, config *websocket.Config, logger *logger.Logger) *WSHandler {
	upgrader := gorillaWs.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &WSHandler{
		wsService: wsService,
		upgrader:  upgrader,
		logger:    logger,
		config:    config,
	}
}

func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		h.logger.Warn("Попытка подключения без идентификатора пользователя")
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Ошибка обновления до WebSocket")
		return
	}

	ctx := h.logger.WithField("userID", userID)
	ctx.Info("Устанавливается новое WebSocket соединение")

	client := websocket.NewClient(conn, userID, h.config, ctx)

	h.wsService.RegisterClient(userID, client)

	client.StartListening(h.wsService.UnregisterClient)
}
