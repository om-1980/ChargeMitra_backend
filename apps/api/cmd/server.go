package cmd

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/om-1980/ChargeMitra_backend/apps/api/routes"
	"github.com/om-1980/ChargeMitra_backend/configs"
	"github.com/om-1980/ChargeMitra_backend/pkg/logger"
)

type Server struct {
	Config *configs.Config
	Logger *logger.Logger
	DB     *pgxpool.Pool
	Router *gin.Engine
}

func NewServer(cfg *configs.Config, log *logger.Logger, db *pgxpool.Pool) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	routes.Register(router, db, cfg)

	return &Server{
		Config: cfg,
		Logger: log,
		DB:     db,
		Router: router,
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf(":%s", s.Config.APIPort)
	s.Logger.Infof("starting API server on %s", address)
	return s.Router.Run(address)
}