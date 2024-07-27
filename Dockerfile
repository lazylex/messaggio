FROM golang:alpine AS builder

WORKDIR /build

COPY . .

# Без использования прокси не скачиваются некоторые библиотеки
RUN go env -w GOPROXY=https://goproxy.cn,direct && go build -o main /build/cmd/main.go

FROM alpine

# Порт для Prometheus
ARG PROMETHEUS_PORT=9323
# Порт HTTP сервера
ARG HTTP_PORT=8897

LABEL authors="lex"

WORKDIR /app

COPY --from=builder /build/main /app/main
COPY --from=builder /build/config/production.yaml /app/production.yaml

EXPOSE $HTTP_PORT
EXPOSE $PROMETHEUS_PORT

ENTRYPOINT ["/app/main", "-config=/app/production.yaml"]