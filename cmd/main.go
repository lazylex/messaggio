package main

import (
	"fmt"
	"github.com/lazylex/messaggio/internal/adapters/http/server"
	"github.com/lazylex/messaggio/internal/adapters/kafka"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/logger"
	prometheusMetrics "github.com/lazylex/messaggio/internal/metrics"
	"github.com/lazylex/messaggio/internal/outbox/naive_implementation/record_outbox"
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

	brokerOutbox := record_outbox.New("brokerOutbox")
	repoOutbox := record_outbox.New("repoOutbox")

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
