.PHONY: build run clean test lint help kafka

BINDIR := bin
SERVERBIN := $(BINDIR)/server
LAUNCHERBIN := $(BINDIR)/launcher
BINARY := notification-service
CONFIG_PATH := $(shell pwd)/config

help:
	@echo "Доступные команды:"
	@echo "  make build      - Компиляция проекта"
	@echo "  make run        - Запуск сервиса"
	@echo "  make clean      - Очистка бинарных файлов"
	@echo "  make test       - Запуск тестов"
	@echo "  make lint       - Проверка кода с помощью golangci-lint"
	@echo "  make kafka      - Запуск Kafka в Docker для локального тестирования"

build:
	@echo "Сборка сервиса уведомлений..."
	@mkdir -p $(BINDIR)
	@go build -o $(SERVERBIN) ./cmd/server
	@go build -o $(LAUNCHERBIN) ./cmd/launcher
	@go build -o $(BINARY) cmd/launcher/main.go
	@echo "Сборка завершена!"

run: build
	@echo "Запуск сервиса уведомлений..."
	@echo "Используемый путь к конфигурации: $(CONFIG_PATH)"
	@CONFIG_PATH=$(CONFIG_PATH) ./$(BINARY)

clean:
	@echo "Очистка бинарных файлов..."
	@rm -rf $(BINDIR) $(BINARY)
	@echo "Очистка завершена!"

test:
	@echo "Запуск тестов..."
	@go test -v ./...

lint:
	@echo "Проверка кода..."
	@golangci-lint run

kafka:
	@echo "Запуск Kafka для локального тестирования..."
	@docker-compose up -d zookeeper kafka kafka-ui 