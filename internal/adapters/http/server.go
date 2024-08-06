package http

import (
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/ports/service"
)

func StartServer(service service.Interface, cfg *config.Config) error {
	if cfg.Env == config.EnvironmentProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	handler := NewHandler(service)

	if cfg.Env != config.EnvironmentLocal {
		tokenMiddleware := NewJWTMiddleware([]byte(cfg.SecureKey))
		router.POST("/msg", tokenMiddleware.CheckJWT(), handler.ProcessMessage)
	} else {
		router.POST("/msg", handler.ProcessMessage)
	}

	router.GET("/statistic", handler.Statistic)
	router.GET("/processed-statistic", handler.ProcessedStatistic)

	return router.Run(fmt.Sprintf("%s:%s", cfg.HttpHost, cfg.HttpPort))
}
