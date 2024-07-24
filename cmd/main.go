package main

import (
	"fmt"
	"github.com/lazylex/messaggio/internal/adapters/http/server"
	"github.com/lazylex/messaggio/internal/adapters/kafka"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/helpers/constants/various"
	"github.com/lazylex/messaggio/internal/logger"
	prometheusMetrics "github.com/lazylex/messaggio/internal/metrics"
	naiveOutbox "github.com/lazylex/messaggio/internal/outbox/naive_implementation/record_outbox"
	"github.com/lazylex/messaggio/internal/outbox/redis_outbox"
	"github.com/lazylex/messaggio/internal/ports/record_outbox"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/lazylex/messaggio/internal/repository/postgresql"
	"github.com/lazylex/messaggio/internal/service"
)

func main() {
	cfg := config.MustLoad()

	slog.SetDefault(logger.MustCreate(cfg.Env, cfg.Instance))
	clearScreen()

	repo := postgresql.MustCreate(cfg.PersistentStorage)
	brokerOutbox, repoOutbox := MustCreateOutboxes(cfg)

	metrics := prometheusMetrics.MustCreate(&cfg.Prometheus)

	domainService := service.MustCreate(repo, brokerOutbox, repoOutbox, cfg.Service, metrics.Service)
	kafka.MustRun(cfg.Kafka, domainService, cfg.Instance)
	httpServer := server.MustCreate(domainService, cfg.HttpServer, cfg.Env, metrics.HTTP)
	httpServer.MustRun()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	fmt.Println() // так красивее, если вывод логов производится в стандартный терминал
	slog.Info(fmt.Sprintf("%s signal received. Shutdown started", sig))

	httpServer.Shutdown()
}

func clearScreen() {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

// MustCreateOutboxes возвращает outbox'ы для временного сохранения сообщений, не отправленных в Kafka и в СУБД. При
// неверно заданной конфигурации (указан несуществующий outbox и т.п.) выдает ошибку в лог и прекращает работу
// приложения.
func MustCreateOutboxes(cfg *config.Config) (brokerOutbox, repoOutbox record_outbox.Interface) {
	switch cfg.Outbox {
	case various.Redis:
		if len(cfg.RedisAddress) == 0 {
			slog.Error("Redis address is empty")
			os.Exit(1)
		}

		redisClient := redis.NewClient(
			&redis.Options{Addr: cfg.RedisAddress, Username: cfg.RedisUser, Password: cfg.RedisPassword, DB: cfg.RedisDB})
		brokerOutbox = redis_outbox.MustCreate(redisClient, "brokerOutbox", cfg.Instance)
		repoOutbox = redis_outbox.MustCreate(redisClient, "repoOutbox", cfg.Instance)
	case various.Naive:
		brokerOutbox = naiveOutbox.New()
		repoOutbox = naiveOutbox.New()
	default:
		slog.Error("Outbox not set")
		os.Exit(1)
	}

	return
}
