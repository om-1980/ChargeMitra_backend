package main

import (
	appcmd "github.com/om-1980/ChargeMitra_backend/apps/ocpp/cmd"
	appconfig "github.com/om-1980/ChargeMitra_backend/apps/ocpp/config"
	"github.com/om-1980/ChargeMitra_backend/internal/db"
	"github.com/om-1980/ChargeMitra_backend/pkg/logger"
)

func main() {
	log := logger.New()

	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("failed to load OCPP config: %v", err)
	}

	pg, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer pg.Close()

	redisClient, err := db.NewRedis(cfg)
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	defer redisClient.Close()

	log.Info("postgres connected successfully")
	log.Info("redis connected successfully")

	server := appcmd.NewServer(cfg, log, pg, redisClient)
	if err := server.Start(); err != nil {
		log.Fatalf("failed to start OCPP server: %v", err)
	}
}