package config

import (
	"fmt"
	"github.com/wb-go/wbf/config"
	"log"
	"net"
	"time"
)

type Config struct {
	DB       DBConfig       `mapstructure:",squash"`
	Server   ServerConfig   `mapstructure:",squash"`
	RabbitMQ RabbitMQConfig `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
	SMTP     SMTPConfig     `mapstructure:",squash"`
	Retry    RetryConfig    `mapstructure:",squash"`
}

type DBConfig struct {
	PgUser          string        `mapstructure:"PGUSER"`
	PgPassword      string        `mapstructure:"PGPASSWORD"`
	PgHost          string        `mapstructure:"PGHOST"`
	PgPort          int           `mapstructure:"PGPORT"`
	PgDatabase      string        `mapstructure:"PGDATABASE"`
	PgSSLMode       string        `mapstructure:"PGSSLMODE"`
	MaxOpenConns    int           `mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"CONN_MAX_LIFETIME"`
}

type ServerConfig struct {
	HTTPPort string `mapstructure:"HTTP_PORT"`
}

type RabbitMQConfig struct {
	User       string        `mapstructure:"RABBIT_USER"`
	Password   string        `mapstructure:"RABBIT_PASSWORD"`
	Host       string        `mapstructure:"RABBIT_HOST"`
	Port       string        `mapstructure:"RABBIT_PORT"`
	Retries    int           `mapstructure:"RETRIES"`
	Pause      time.Duration `mapstructure:"PAUSE"`
	Exchange   string        `mapstructure:"EXCHANGE"`
	RoutingKey string        `mapstructure:"ROUTING_KEY"`
	Queue      string        `mapstructure:"QUEUE"`
	DLQ        string        `mapstructure:"DLQ"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     string `mapstructure:"SMTP_PORT"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	From     string `mapstructure:"FROM"`
}

type RetryConfig struct {
	Attempts int           `mapstructure:"RETRY_ATTEMPTS"`
	Delay    time.Duration `mapstructure:"RETRY_DELAY"`
	Backoff  float64       `mapstructure:"RETRY_BACKOFF"`
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

func (r *RabbitMQConfig) Url() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", r.User, r.Password, r.Host, r.Port)
}

func (r *RedisConfig) Addr() string {
	return net.JoinHostPort(r.Host, r.Port)
}
