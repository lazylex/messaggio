package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/lazylex/messaggio/internal/adapters/http/handlers"
	"github.com/lazylex/messaggio/internal/adapters/http/middleware/jwt"
	"github.com/lazylex/messaggio/internal/adapters/http/middleware/recoverer"
	requestMetrics "github.com/lazylex/messaggio/internal/adapters/http/middleware/request_metrics"
	"github.com/lazylex/messaggio/internal/adapters/http/router"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/helpers/constants/prefixes"
	mi "github.com/lazylex/messaggio/internal/ports/metrics/http"
	"github.com/lazylex/messaggio/internal/service"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"time"
)

// Server структура для обработки http-запросов к приложению.
type Server struct {
	cfg     *config.HttpServer // Конфигурация http сервера
	srv     *http.Server       // Структура с параметрами сервера
	mux     *http.ServeMux     // Мультиплексор http запросов
	service *service.Service   // Структура, реализующая логику приложения
}

// MustCreate создает и возвращает http-сервер.
func MustCreate(domainService *service.Service, cfg config.HttpServer, env string, metrics mi.MetricsInterface) *Server {
	mux := http.NewServeMux()
	server := &Server{mux: mux, service: domainService, cfg: &cfg}

	h := handlers.New(domainService, cfg.RequestTimeout)

	router.AssignPathToHandler("/msg", server.mux, h.ProcessMessage)
	router.AssignPathToHandler("/statistic", server.mux, h.Statistic)
	router.AssignPathToHandler("/processed-statistic", server.mux, h.ProcessedStatistic)

	server.srv = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", server.cfg.HttpPort),
		Handler:      server.mux,
		ReadTimeout:  server.cfg.ReadTimeout,
		WriteTimeout: server.cfg.WriteTimeout,
		IdleTimeout:  server.cfg.IdleTimeout,
	}

	if cfg.EnableProfiler {
		if cfg.WriteTimeout <= time.Second*30 {
			slog.Warn("standard profile duration exceeds server's WriteTimeout")
		}

		router.AssignPathToHandler(prefixes.PPROFPrefix, server.mux, pprof.Index)
		router.AssignPathToHandler(prefixes.PPROFPrefix+"cmdline", server.mux, pprof.Cmdline)
		router.AssignPathToHandler(prefixes.PPROFPrefix+"profile", server.mux, pprof.Profile)
		router.AssignPathToHandler(prefixes.PPROFPrefix+"symbol", server.mux, pprof.Symbol)
		router.AssignPathToHandler(prefixes.PPROFPrefix+"trace", server.mux, pprof.Trace)
	}

	if env != config.EnvironmentLocal {
		tokenMiddleware := jwt.New([]byte(cfg.SecureKey))
		server.srv.Handler = tokenMiddleware.CheckJWT(server.mux)
	}

	metricsMiddleware := requestMetrics.New(metrics)
	server.srv.Handler = recoverer.Recoverer(server.srv.Handler)
	server.srv.Handler = metricsMiddleware.BeforeHandle(server.srv.Handler)
	server.srv.Handler = metricsMiddleware.AfterHandle(server.srv.Handler)

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
