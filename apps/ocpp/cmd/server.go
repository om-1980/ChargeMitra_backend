package cmd

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	appconfig "github.com/om-1980/ChargeMitra_backend/apps/ocpp/config"
	"github.com/om-1980/ChargeMitra_backend/apps/ocpp/ws"
	"github.com/om-1980/ChargeMitra_backend/internal/ocppcore"
	"github.com/om-1980/ChargeMitra_backend/pkg/logger"
)

type Server struct {
	Config *appconfig.Config
	Logger *logger.Logger
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Router *gin.Engine
}

func NewServer(cfg *appconfig.Config, log *logger.Logger, db *pgxpool.Pool, redisClient *redis.Client) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	hub := ocppcore.NewHub()
	ocppService := ocppcore.NewService(db, redisClient, hub)
	wsServer := ws.NewServer(log, ocppService)
	apiHandler := ocppcore.NewAPIHandler(ocppService)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "ocpp",
			"status":  "ok",
		})
	})

	router.GET("/ws/:ocppId", wsServer.HandleChargePoint)

	protected := router.Group("/")
	protected.Use(ocppcore.RequireControlToken())
	{
		protected.GET("/connections", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"connections": hub.List(),
			})
		})

		protected.POST("/control/remote-start", apiHandler.RemoteStart)
		protected.POST("/control/remote-stop", apiHandler.RemoteStop)
	}

	return &Server{
		Config: cfg,
		Logger: log,
		DB:     db,
		Redis:  redisClient,
		Router: router,
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf(":%s", s.Config.OCPPPort)
	s.Logger.Infof("starting OCPP server on %s", address)
	return s.Router.Run(address)
}