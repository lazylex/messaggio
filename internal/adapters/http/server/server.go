package server

import (
	"context"
	"errors"
	"github.com/lazylex/messaggio/internal/adapters/http/handlers"
	"github.com/lazylex/messaggio/internal/adapters/http/middleware/recoverer"
	"github.com/lazylex/messaggio/internal/adapters/http/router"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/service"
	"log/slog"
	"net/http"
	"os"
)

// Server структура для обработки http-запросов к приложению.
type Server struct {
	cfg     *config.HttpServer // Конфигурация http сервера
	srv     *http.Server       // Структура с параметрами сервера
	mux     *http.ServeMux     // Мультиплексор http запросов
	service *service.Service   // Структура, реализующая логику приложения
}

// MustCreate создает и возвращает http-сервер.
func MustCreate(domainService *service.Service, cfg config.HttpServer) *Server {
	mux := http.NewServeMux()
	server := &Server{mux: mux, service: domainService, cfg: &cfg}

	h := handlers.New(domainService, cfg.RequestTimeout)

	router.AssignPathToHandler("/msg", server.mux, h.ProcessMessage)
	router.AssignPathToHandler("/statistic", server.mux, h.Statistic)

	server.srv = &http.Server{
		Addr:         server.cfg.Address,
		Handler:      server.mux,
		ReadTimeout:  server.cfg.ReadTimeout,
		WriteTimeout: server.cfg.WriteTimeout,
		IdleTimeout:  server.cfg.IdleTimeout,
	}

	server.srv.Handler = recoverer.Recoverer(server.srv.Handler)

	return server
}

// MustRun производит запуск сервера в отдельной go-рутине. В случае ошибки останавливает работу приложения.
func (s *Server) MustRun() {
	go func() {
		slog.Info("start http server on " + s.srv.Addr)
		err := s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server err: startup error. Initial error: " + err.Error())
			os.Exit(1)
		}
	}()
}

// Shutdown производит остановку сервера.
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		slog.Error("failed to gracefully shutdown http server")
	} else {
		slog.Info("gracefully shut down http server")
	}
}
