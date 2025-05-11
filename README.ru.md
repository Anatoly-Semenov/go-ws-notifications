# go-ws-notifications

Высокопроизводительный сервис на Go для отправки уведомлений пользователям через WebSocket в реальном времени.

## Особенности

- Получение уведомлений из Kafka (топик 'notifications.web') и отправка их клиентам через WebSocket
- Аутентификация пользователей
- TLS шифрование
- Масштабируемость и высокая производительность
- Prometheus метрики для мониторинга
- Контекстуальное логирование с использованием uber-go/zap
- Валидация данных с помощью go-playground/validator

## Архитектура

Проект реализован с учетом принципов Clean Architecture:

- Domain layer - бизнес-сущности и интерфейсы
- Application layer - реализация бизнес-логики
- Infrastructure layer - внешние зависимости:
  - Kafka для получения уведомлений
  - WebSocket для отправки клиентам
  - HTTP для API
  - Repository для хранения данных

## Установка

```bash
git clone https://github.com/anatoly_dev/go-ws-notifications.git
cd go-ws-notifications
go mod download
```

## Конфигурация

Настройки сервиса содержатся в файле `config/config.yaml`. Пример конфигурации:

```yaml
server:
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  metrics_port: 9090

kafka:
  brokers: ["localhost:9092"]
  topic: "notifications.web"
  group_id: "notification-service"
  auto_offset_reset: "earliest"

websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  pong_wait: 60s
  ping_period: 54s
  max_message_size: 512000

tls:
  enabled: false
  cert_file: "certs/server.crt"
  key_file: "certs/server.key"
```

## Запуск сервиса

### Запуск локально с Kafka

```bash
# Сначала запустить Kafka
make kafka

# Затем запустить сервис
make run
```

### Запуск через Docker

```bash
# Запустить все сервисы включая Kafka, Prometheus, Grafana и т.д.
docker-compose up
```

## Доступные эндпоинты

- WebSocket API: `ws://localhost:8080/ws`
- Проверка состояния: `http://localhost:8080/health`
- Метрики Prometheus: `http://localhost:9090/metrics`

При запуске через Docker также доступны:
- Kafka UI: `http://localhost:8090`
- Grafana: `http://localhost:3000` (логин/пароль: admin/admin)
- Prometheus: `http://localhost:9091`

## Структура проекта

```
/cmd
  /server    - Основной сервер приложения
  /launcher  - Запускатель для управления сервером
/config      - Конфигурация приложения
/internal    - Внутренняя реализация
  /application     - Бизнес-логика приложения
  /domain          - Модели и интерфейсы домена
  /infrastructure  - Реализация инфраструктуры
    /http          - HTTP сервер
    /kafka         - Интеграция с Kafka
    /repository    - Хранилища данных
    /websocket     - Реализация WebSocket
/pkg         - Общие пакеты
```

## Формат уведомлений

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user123",
  "type": "system",
  "title": "Новое сообщение",
  "content": "У вас новое сообщение от администратора",
  "is_read": false,
  "created_at": "2024-01-01T12:00:00Z",
  "priority": 3
}
```

## Метрики

Prometheus метрики доступны по адресу `http://localhost:9090/metrics`
