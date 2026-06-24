package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Yertore/sub-aggregator/internal/config"
	"github.com/Yertore/sub-aggregator/internal/handler"
	"github.com/Yertore/sub-aggregator/internal/repository"
	"github.com/Yertore/sub-aggregator/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		slog.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to database", "host", cfg.DBHost, "db", cfg.DBName)

	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/cost", h.TotalCost)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	slog.Info("starting server", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
