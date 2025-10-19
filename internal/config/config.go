package config

import (
	"github.com/wb-go/wbf/config"
	"log"
)

type Config struct {
	DB     DBConfig     `mapstructure:",squash"`
	Server ServerConfig `mapstructure:",squash"`
}

type DBConfig struct {
	PgUser     string `mapstructure:"PGUSER"`
	PgPassword string `mapstructure:"PGPASSWORD"`
	PgHost     string `mapstructure:"PGHOST"`
	PgPort     int    `mapstructure:"PGPORT"`
	PgDatabase string `mapstructure:"PGDATABASE"`
	PgSSLMode  string `mapstructure:"PGSSLMODE"`
}

type ServerConfig struct {
	HTTPPort string `mapstructure:"HTTP_PORT"`
}

func MustLoad() *Config {

	c := config.New()
	if err := c.Load(".env", ".env", ""); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var cfg Config

	if err := c.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &cfg
}
