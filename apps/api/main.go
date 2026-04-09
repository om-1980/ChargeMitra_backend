package main

import (
	"github.com/om-1980/ChargeMitra_backend/apps/api/cmd"
	"github.com/om-1980/ChargeMitra_backend/configs"
	"github.com/om-1980/ChargeMitra_backend/internal/db"
	"github.com/om-1980/ChargeMitra_backend/pkg/logger"
)

func main() {
	log := logger.New()

	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	
	log.Infof("DB CONFIG => host=%s port=%s user=%s db=%s password=%q",
	cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBPassword)

	pg, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pg.Close()

	log.Info("postgres connected successfully")

	server := cmd.NewServer(cfg, log, pg)

	if err := server.Start(); err != nil {
		log.Fatalf("failed to start API server: %v", err)
	}
}