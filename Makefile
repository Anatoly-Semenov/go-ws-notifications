.PHONY: build run clean test lint help

BINDIR := bin
SERVERBIN := $(BINDIR)/server
LAUNCHERBIN := $(BINDIR)/launcher
BINARY := notification-service

help:
	@echo "Доступные команды:"
	@echo "  make build      - Компиляция проекта"
	@echo "  make run        - Запуск сервиса"
	@echo "  make clean      - Очистка бинарных файлов"
	@echo "  make test       - Запуск тестов"
	@echo "  make lint       - Проверка кода с помощью golangci-lint"

build:
	@echo "Сборка сервиса уведомлений..."
	@mkdir -p $(BINDIR)
	@go build -o $(SERVERBIN) ./cmd/server
	@go build -o $(LAUNCHERBIN) ./cmd/launcher
	@go build -o $(BINARY) cmd/launcher/main.go
	@echo "Сборка завершена!"

run: build
	@echo "Запуск сервиса уведомлений..."
	@./$(BINARY)

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