# go-ws-notifications

High-performance Go service for sending real-time notifications to users via WebSocket.

## Features

- Consumes notifications from Kafka (topic: 'notifications.web') and sends them to clients via WebSocket
- User authentication
- TLS encryption
- Scalable and high-performance
- Prometheus metrics for monitoring
- Contextual logging using uber-go/zap
- Data validation with go-playground/validator

## Architecture

The project follows the principles of Clean Architecture:

- Domain layer - business entities and interfaces
- Application layer - business logic implementation
- Infrastructure layer - external dependencies:
  - Kafka for receiving notifications
  - WebSocket for sending notifications to clients
  - HTTP for API
  - Repository for data storage

## Installation

```bash
git clone https://github.com/anatoly_dev/go-ws-notifications.git
cd go-ws-notifications
go mod download
```

## Configuration

Service settings are located in the `config/config.yaml` file. Example configuration:

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

## Running the Service

### Local Run with Kafka

```bash
# Start Kafka first
make kafka

# Then start the service
make run
```

### Run with Docker

```bash
# Start all services including Kafka, Prometheus, Grafana, etc.
docker-compose up
```

## Available Endpoints

- WebSocket API: `ws://localhost:8080/ws`
- Health Check: `http://localhost:8080/health`
- Prometheus Metrics: `http://localhost:9090/metrics`

When running with Docker, also available:
- Kafka UI: `http://localhost:8090`
- Grafana: `http://localhost:3000` (login/password: admin/admin)
- Prometheus: `http://localhost:9091`

## Project Structure

```
/cmd
  /server    - Main application server
  /launcher  - Launcher for managing the server
/config      - Application configuration
/internal    - Internal implementation
  /application     - Application business logic
  /domain          - Domain models and interfaces
  /infrastructure  - Infrastructure implementation
    /http          - HTTP server
    /kafka         - Kafka integration
    /repository    - Data storage
    /websocket     - WebSocket implementation
/pkg         - Shared packages
```

## Notification Format

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user123",
  "type": "system",
  "title": "New Message",
  "content": "You have a new message from the administrator",
  "is_read": false,
  "created_at": "2024-01-01T12:00:00Z",
  "priority": 3
}
```

## Metrics

Prometheus metrics are available at `http://localhost:9090/metrics`