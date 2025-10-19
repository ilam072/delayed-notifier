package main

import (
	"delayed-notifier/internal/config"
	"delayed-notifier/pkg/db"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()
	cfg := config.MustLoad()
	db, err := db.OpenDB(cfg.DB)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	_ = db
}
