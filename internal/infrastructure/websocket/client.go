package websocket

import (
	"sync"
	"time"

	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn       *websocket.Conn
	send       chan []byte
	userID     string
	logger     *logger.Logger
	config     *Config
	isClosed   bool
	closeMutex sync.Mutex
}

func NewClient(conn *websocket.Conn, userID string, config *Config, logger *logger.Logger) *Client {
	return &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		logger:   logger.WithField("userID", userID),
		config:   config,
		isClosed: false,
	}
}

func (c *Client) StartListening(unregisterFunc func(userID string)) {
	go c.writePump()
	go c.readPump(unregisterFunc)
}

func (c *Client) Send(message []byte) error {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	if c.isClosed {
		return nil
	}

	select {
	case c.send <- message:
		return nil
	default:
		c.logger.Warn("Буфер сообщений клиента переполнен")
		return c.closeUnsafe()
	}
}

func (c *Client) Close() error {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	return c.closeUnsafe()
}

func (c *Client) closeUnsafe() error {
	if c.isClosed {
		return nil
	}

	c.isClosed = true

	close(c.send)

	message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	_ = c.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))

	return c.conn.Close()
}

func (c *Client) readPump(unregisterFunc func(userID string)) {
	defer func() {
		c.logger.Info("Завершение чтения сообщений от клиента")
		c.Close()
		unregisterFunc(c.userID)
	}()

	c.conn.SetReadLimit(c.config.MaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.config.PongWait) * time.Second))

	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.config.PongWait) * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.WithError(err).Error("Неожиданная ошибка при чтении WebSocket сообщения")
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(time.Duration(c.config.PingPeriod) * time.Second)
	defer func() {
		c.logger.Info("Завершение отправки сообщений клиенту")
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))

			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.logger.WithError(err).Error("Ошибка получения writer для WebSocket")
				return
			}

			_, err = w.Write(message)
			if err != nil {
				c.logger.WithError(err).Error("Ошибка записи сообщения в WebSocket")
				return
			}

			if err := w.Close(); err != nil {
				c.logger.WithError(err).Error("Ошибка закрытия writer для WebSocket")
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.WithError(err).Error("Ошибка отправки ping-сообщения")
				return
			}
		}
	}
}
