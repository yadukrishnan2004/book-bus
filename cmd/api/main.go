package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.New(ctx, cfg.DSN())
	if err != nil {
		slog.Error("could not connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("connected to PostgreSQL", "host", cfg.DBHost, "db", cfg.DBName)

	// 4. Build layers (innermost → outermost)
	userRepo := repository.NewUserRepository(pool)
	busRepo := repository.NewBusRepository(pool)
	routeRepo := repository.NewRouteRepository(pool)
	scheduleRepo := repository.NewScheduleRepository(pool)
	bookingRepo := repository.NewBookingRepository(pool)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	busSvc := service.NewBusService(busRepo)
	routeSvc := service.NewRouteService(routeRepo)
	scheduleSvc := service.NewScheduleService(scheduleRepo)
	bookingSvc := service.NewBookingService(bookingRepo, scheduleRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	busHandler := handler.NewBusHandler(busSvc)
	routeHandler := handler.NewRouteHandler(routeSvc)
	scheduleHandler := handler.NewScheduleHandler(scheduleSvc)
	bookingHandler := handler.NewBookingHandler(bookingSvc)

	// 5. Start HTTP server
	appServer := server.New(
		pool,
		authSvc,
		authHandler,
		busHandler,
		routeHandler,
		scheduleHandler,
		bookingHandler,
	)
	httpSrv := appServer.HTTPServer(cfg.Port)

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 6. Graceful shutdown listening on SIGINT and SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		slog.Error("fatal server error", "error", err)
		os.Exit(1)
	case sig := <-shutdown:
		slog.Info("start graceful shutdown", "signal", sig.String())

		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		if err := httpSrv.Shutdown(ctxShutdown); err != nil {
			slog.Error("server forced to shutdown", "error", err)
			httpSrv.Close()
		}

		slog.Info("server shutdown complete")
	}
}