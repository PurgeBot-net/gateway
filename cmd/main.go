package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"github.com/PurgeBot-net/common/log"
	"github.com/PurgeBot-net/database"
	"github.com/PurgeBot-net/gateway/config"
	"github.com/PurgeBot-net/gateway/internal/events"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("load config: " + err.Error())
	}

	logger, err := log.New(cfg.LogLevel, cfg.LogJSON)
	if err != nil {
		panic("create logger: " + err.Error())
	}
	logger = log.WithSentry(logger, cfg.SentryDSN)
	defer logger.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := database.New(ctx, cfg.DatabaseURL())
	if err != nil {
		logger.Fatal("connect database", zap.Error(err))
	}
	defer db.Close()

	gw := events.NewGateway(cfg, logger, db)
	if err := gw.Start(ctx); err != nil {
		logger.Fatal("gateway stopped", zap.Error(err))
	}
}
