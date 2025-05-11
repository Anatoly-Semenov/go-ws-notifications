package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Kafka     KafkaConfig     `mapstructure:"kafka"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	TLS       TLSConfig       `mapstructure:"tls"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	MetricsPort  int           `mapstructure:"metrics_port"`
}

type KafkaConfig struct {
	Brokers         []string `mapstructure:"brokers"`
	Topic           string   `mapstructure:"topic"`
	GroupID         string   `mapstructure:"group_id"`
	AutoOffsetReset string   `mapstructure:"auto_offset_reset"`
}

type WebSocketConfig struct {
	ReadBufferSize  int           `mapstructure:"read_buffer_size"`
	WriteBufferSize int           `mapstructure:"write_buffer_size"`
	PongWait        time.Duration `mapstructure:"pong_wait"`
	PingPeriod      time.Duration `mapstructure:"ping_period"`
	MaxMessageSize  int64         `mapstructure:"max_message_size"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("ошибка разбора файла конфигурации: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {

	if len(config.Kafka.Brokers) == 0 {
		return fmt.Errorf("не указаны адреса брокеров Kafka")
	}

	if config.Kafka.Topic == "" {
		return fmt.Errorf("не указан топик Kafka")
	}

	if config.Server.Port <= 0 {
		config.Server.Port = 8080
	}

	if config.Server.MetricsPort <= 0 {
		config.Server.MetricsPort = 9090
	}

	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 15 * time.Second
	}

	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 15 * time.Second
	}

	if config.WebSocket.ReadBufferSize <= 0 {
		config.WebSocket.ReadBufferSize = 1024
	}

	if config.WebSocket.WriteBufferSize <= 0 {
		config.WebSocket.WriteBufferSize = 1024
	}

	if config.WebSocket.PongWait == 0 {
		config.WebSocket.PongWait = 60 * time.Second
	}

	if config.WebSocket.PingPeriod == 0 {

		config.WebSocket.PingPeriod = (config.WebSocket.PongWait * 9) / 10
	}

	if config.WebSocket.MaxMessageSize <= 0 {
		config.WebSocket.MaxMessageSize = 512000
	}

	return nil
}
