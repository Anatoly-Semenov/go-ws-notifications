FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/server ./cmd/server

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

RUN adduser -D -g '' appuser

COPY --from=builder /app/bin/server /app/bin/server
COPY --chown=appuser:appuser config /app/config

RUN mkdir -p /app/certs && \
    chown -R appuser:appuser /app/certs

USER appuser

EXPOSE 8080 9090

CMD ["/app/bin/server"] 