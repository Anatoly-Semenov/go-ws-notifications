package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/anatoly_dev/go-ws-notifications/config"
	"github.com/anatoly_dev/go-ws-notifications/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	server        *http.Server
	metricsServer *http.Server
	logger        *logger.Logger
	wsHandler     *WSHandler
}

func NewServer(cfg *config.Config, wsHandler *WSHandler, logger *logger.Logger) *Server {
	router := http.NewServeMux()

	router.HandleFunc("/ws", wsHandler.HandleConnection)

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if cfg.TLS.Enabled {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	metricsRouter := http.NewServeMux()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.MetricsPort),
		Handler: metricsRouter,
	}

	return &Server{
		server:        server,
		metricsServer: metricsServer,
		logger:        logger,
		wsHandler:     wsHandler,
	}
}

func (s *Server) Start() error {
	// Запуск сервера метрик в отдельной горутине
	go func() {
		s.logger.WithField("port", s.metricsServer.Addr).Info("Запуск сервера метрик")
		if err := s.metricsServer.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				s.logger.Info("Сервер метрик корректно остановлен")
			} else {
				s.logger.WithError(err).Error("Ошибка запуска сервера метрик")
			}
		}
	}()

	// Небольшая задержка, чтобы сервер метрик успел запуститься
	time.Sleep(100 * time.Millisecond)

	s.logger.WithField("port", s.server.Addr).Info("Запуск HTTP сервера")

	var err error
	if s.server.TLSConfig != nil {
		err = s.server.ListenAndServeTLS("", "")
	} else {
		err = s.server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Остановка HTTP сервера")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Ошибка остановки HTTP сервера")
		return err
	}

	if err := s.metricsServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Ошибка остановки сервера метрик")
		return err
	}

	s.logger.Info("HTTP сервер успешно остановлен")
	return nil
}
