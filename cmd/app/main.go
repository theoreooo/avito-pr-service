package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avito-pr-service/internal/api"
	"avito-pr-service/internal/config"
	"avito-pr-service/internal/lib/logger/sl"
	"avito-pr-service/internal/repository/postgres"
	"avito-pr-service/internal/server/handler"
	"avito-pr-service/internal/service"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	log.Info("starting app", slog.Any("cfg", cfg))

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storageCtx, storageCancel := context.WithTimeout(ctx, 10*time.Second)
	defer storageCancel()

	storage, err := postgres.New(storageCtx, connStr)
	if err != nil {
		log.Error("failed to connect to postgres", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	log.Info("connected to postgres")

	prRepo := postgres.NewPRRepository(storage.Pool())
	userRepo := postgres.NewUserRepository(storage.Pool())
	teamRepo := postgres.NewTeamRepository(storage.Pool(), userRepo)
	statRepo := postgres.NewStatisticsRepository(storage.Pool())

	prService := service.NewPRService(prRepo)
	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo, userRepo, prRepo)
	statService := service.NewStatisticsService(statRepo)

	prHandler := handler.NewPRHandler(prService)
	userHandler := handler.NewUserHandler(userService, prService)
	teamHandler := handler.NewTeamHandler(teamService)
	statHandler := handler.NewStatisticsHandler(statService)

	apiHandler := handler.NewAPIHandler(prHandler, userHandler, teamHandler, statHandler)
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	api.HandlerFromMux(apiHandler, r)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/openapi.json"),
	))

	r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./openapi.yaml")
	})

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", sl.Err(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Info("service ready, waiting for stop signal")
	<-stop

	log.Info("shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", sl.Err(err))
	}

	log.Info("service stopped")
}
