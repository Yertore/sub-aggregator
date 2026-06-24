package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Yertore/sub-aggregator/internal/config"
	"github.com/Yertore/sub-aggregator/internal/handler"
	"github.com/Yertore/sub-aggregator/internal/repository"
	"github.com/Yertore/sub-aggregator/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("database ping failed:", err)
	}

	log.Println("connected to database")

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

	log.Printf("Starting server on :%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatal(err)
	}
}
