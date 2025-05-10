package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anatoly_dev/go-ws-notifications/config"
	"github.com/anatoly_dev/go-ws-notifications/internal/application"
	"github.com/anatoly_dev/go-ws-notifications/internal/infrastructure/http"
	"github.com/anatoly_dev/go-ws-notifications/internal/infrastructure/kafka"
	"github.com/anatoly_dev/go-ws-notifications/internal/infrastructure/repository"
	"github.com/anatoly_dev/go-ws-notifications/internal/infrastructure/websocket"
	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
)

type App struct {
	cfg              *config.Config
	logger           *logger.Logger
	notificationRepo *repository.MemoryRepository
	wsService        *websocket.Service
	notificationSvc  *application.NotificationService
	server           *http.Server
	kafkaConsumer    *kafka.Consumer
}

func main() {
	app := NewApp()
	app.Run()
}

func NewApp() *App {
	cfg, err := config.LoadConfig("./conf")
	if err != nil {
		panic("Ошибка загрузки конфигурации: " + err.Error())
	}

	appLogger, err := logger.NewLogger("info", false)
	if err != nil {
		panic("Ошибка инициализации логгера: " + err.Error())
	}

	appLogger.Info("Инициализация сервиса уведомлений")

	return &App{
		cfg:    cfg,
		logger: appLogger,
	}
}

func (a *App) InitializeServices() {
	a.notificationRepo = repository.NewMemoryRepository(a.logger)

	wsConfig := &websocket.Config{
		ReadBufferSize:  a.cfg.WebSocket.ReadBufferSize,
		WriteBufferSize: a.cfg.WebSocket.WriteBufferSize,
		PongWait:        int(a.cfg.WebSocket.PongWait.Seconds()),
		PingPeriod:      int(a.cfg.WebSocket.PingPeriod.Seconds()),
		MaxMessageSize:  a.cfg.WebSocket.MaxMessageSize,
	}
	a.wsService = websocket.NewService(wsConfig, a.logger)

	a.notificationSvc = application.NewNotificationService(a.notificationRepo, a.wsService, a.logger)

	wsHandler := http.NewWSHandler(a.wsService, wsConfig, a.logger)

	a.server = http.NewServer(a.cfg, wsHandler, a.logger)
}

func (a *App) InitializeKafka() error {
	kafkaConfig := &kafka.ConsumerConfig{
		Brokers:         a.cfg.Kafka.Brokers,
		GroupID:         a.cfg.Kafka.GroupID,
		AutoOffsetReset: a.cfg.Kafka.AutoOffsetReset,
	}

	var err error
	a.kafkaConsumer, err = kafka.NewConsumer(kafkaConfig, a.logger)
	if err != nil {
		return err
	}

	kafkaHandler := application.NewKafkaHandler(a.notificationSvc, a.logger)

	if err := a.kafkaConsumer.Subscribe(a.cfg.Kafka.Topic, kafkaHandler.HandleMessage); err != nil {
		return err
	}

	return nil
}

func (a *App) StartServer() {
	go func() {
		if err := a.server.Start(); err != nil {
			a.logger.WithError(err).Error("Ошибка запуска HTTP сервера")
		}
	}()

	a.logger.Info("Сервис запущен и готов обрабатывать запросы")
}

func (a *App) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	a.logger.WithField("signal", sig.String()).Info("Получен сигнал остановки")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Stop(ctx); err != nil {
		a.logger.WithError(err).Error("Ошибка остановки HTTP сервера")
	}

	a.logger.Info("Сервис успешно остановлен")
}

func (a *App) Cleanup() {
	if a.kafkaConsumer != nil {
		a.kafkaConsumer.Close()
	}

	if a.logger != nil {
		a.logger.Sync()
	}
}

func (a *App) Run() {
	defer a.Cleanup()

	a.InitializeServices()

	if err := a.InitializeKafka(); err != nil {
		a.logger.WithError(err).Fatal("Ошибка инициализации Kafka")
		return
	}

	a.StartServer()
	a.WaitForShutdown()
}
