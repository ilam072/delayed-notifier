package main

import (
	"context"
	"delayed-notifier/internal/config"
	"delayed-notifier/internal/notification/cache"
	"delayed-notifier/internal/notification/rabbitmq/handler"
	"delayed-notifier/internal/notification/rabbitmq/notifier"
	"delayed-notifier/internal/notification/repo/postgres"
	"delayed-notifier/internal/notification/rest"
	"delayed-notifier/internal/notification/senders"
	"delayed-notifier/internal/notification/service"
	"delayed-notifier/internal/notification/worker"
	"delayed-notifier/internal/validator"
	"delayed-notifier/pkg/clients/email"
	"delayed-notifier/pkg/db"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const workersCount = 5

func main() {
	// Initialize logger
	zlog.Init()

	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Initialize config
	cfg := config.MustLoad()

	// Connect to DB
	DB, err := db.OpenDB(cfg.DB)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	// Connect to RabbitMQ
	conn, err := rabbitmq.Connect(cfg.RabbitMQ.Addr(), cfg.RabbitMQ.Retries, cfg.RabbitMQ.Pause)
	if err != nil {
		if err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to connect to RabbitMQ")
		}
	}

	channel, err := conn.Channel()
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to open server channel")
	}

	// Create notifier.Notifier
	notifierOpts := notifier.Opts{
		Exchange:   cfg.RabbitMQ.Exchange,
		RoutingKey: cfg.RabbitMQ.RoutingKey,
		Queue:      cfg.RabbitMQ.Queue,
		DLQ:        cfg.RabbitMQ.DLQ,
	}
	notifierr, err := notifier.New(channel, notifierOpts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create notifier")
	}

	// Connect to Redis
	redisClient := redis.New(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)

	if err = redisClient.Ping(ctx).Err(); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to ping redis")
	}

	// Initialize email client
	emailClient := email.New(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From)

	// Initialize notification senders
	notificationSenders := senders.New(emailClient)

	// Initialize notification repo
	repo := postgres.New(DB)

	// Initialize notification cache
	c := cache.New(redisClient)

	// Initialize notification service
	notificationService := service.NewNotification(repo, notifierr, c, notificationSenders)

	// Initialize retry strategy
	strategy := retry.Strategy{
		Attempts: cfg.Retry.Attempts,
		Delay:    cfg.Retry.Delay,
		Backoff:  cfg.Retry.Backoff,
	}

	// Initialize notification validator
	notificationValidator := validator.New()

	// Initialize notification handlers
	httpHandler := rest.New(notificationService, notificationValidator, strategy)
	msgsHandler := handler.New(notificationService)

	// Init and start workers
	workers := worker.NewWorkerPool(notifierr, msgsHandler, notificationService, workersCount)
	go workers.Start(ctx, strategy)

	// Initialize Gin engine
	engine := ginext.New("")
	engine.Use(ginext.Logger())
	engine.Use(ginext.Recovery())

	apiGroup := engine.Group("/api/notify")
	apiGroup.POST("/", httpHandler.CreateNotification)
	apiGroup.GET("/:id", httpHandler.GetNotificationStatus)
	apiGroup.DELETE("/:id", httpHandler.CancelNotification)

	// Initialize and start http server
	server := &http.Server{
		Addr:    cfg.Server.HTTPPort,
		Handler: engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to listen start http server")
		}
	}()

	<-ctx.Done()

	// Graceful shutdown
	withTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(withTimeout); err != nil {
		zlog.Logger.Error().Err(err).Msg("server shutdown failed")
	}

	if err := DB.Master.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close master database")
	}

	if err := channel.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close channel")
	}

	if err := conn.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close RabbitMQ conn")
	}
}
