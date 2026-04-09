package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName    string
	AppEnv     string
	APIPort    string
	OCPPPort   string
	WorkerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisHost string
	RedisPort string

	JWTSecret string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppName:    getEnv("APP_NAME", "ChargeMitra"),
		AppEnv:     getEnv("APP_ENV", "development"),
		APIPort:    getEnv("API_PORT", "8080"),
		OCPPPort:   getEnv("OCPP_PORT", "8081"),
		WorkerPort: getEnv("WORKER_PORT", "8082"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "charge_mitra"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		JWTSecret: getEnv("JWT_SECRET", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var missing []string

	if c.AppName == "" {
		missing = append(missing, "APP_NAME")
	}
	if c.APIPort == "" {
		missing = append(missing, "API_PORT")
	}
	if c.DBHost == "" {
		missing = append(missing, "DB_HOST")
	}
	if c.DBPort == "" {
		missing = append(missing, "DB_PORT")
	}
	if c.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if c.DBName == "" {
		missing = append(missing, "DB_NAME")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env values: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (c *Config) DatabaseURL() string {
	if c.DBPassword == "" {
		return fmt.Sprintf(
			"postgres://%s@%s:%s/%s?sslmode=disable",
			c.DBUser,
			c.DBHost,
			c.DBPort,
			c.DBName,
		)
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}