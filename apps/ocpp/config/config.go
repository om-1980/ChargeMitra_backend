package config

import "github.com/om-1980/ChargeMitra_backend/configs"

type Config = configs.Config

func Load() (*Config, error) {
	return configs.Load()
}