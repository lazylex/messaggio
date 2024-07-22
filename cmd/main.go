package main

import (
	"fmt"
	"github.com/lazylex/messaggio/internal/adapters/http/server"
	"github.com/lazylex/messaggio/internal/adapters/kafka"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/logger"
	"github.com/lazylex/messaggio/internal/outbox/naive_implementation/id_outbox"
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
	statusOutbox := id_outbox.New("statusOutbox")
	brokerOutbox := record_outbox.New("brokerOutbox")
	repoOutbox := record_outbox.New("repoOutbox")
	domainService := service.MustCreate(repo, statusOutbox, brokerOutbox, repoOutbox, cfg.Service)
	kafka.MustRun(cfg.Kafka, domainService, cfg.Instance)
	httpServer := server.MustCreate(domainService, cfg.HttpServer, cfg.Env)
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
