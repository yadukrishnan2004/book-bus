package main

import (
	"context"
	"log/slog"
	"os"

	"book-bus/internal/config"
	"book-bus/internal/db"
	"book-bus/internal/handler"
	"book-bus/internal/logger"
	"book-bus/internal/repository"
	"book-bus/internal/server"
	"book-bus/internal/service"
)

func main() {
	// 1. Init structured logger
	logger.Init()

	// 2. Load configuration
	cfg := config.Load()

	// 3. Connect to database
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DSN())
	if err != nil {
		slog.Error("could not connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("connected to PostgreSQL", "host", cfg.DBHost, "db", cfg.DBName)

	// 4. Build layers (innermost → outermost)
	busRepo    := repository.NewBusRepository(pool)
	busSvc     := service.NewBusService(busRepo)
	busHandler := handler.NewBusHandler(busSvc)

	// 5. Start server
	srv := server.New(pool, busHandler)

	slog.Info("server starting", "port", cfg.Port)
	if err := srv.Run(cfg.Port); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}