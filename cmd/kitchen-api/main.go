package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "avito-kitchen/internal/delivery/http"
	"avito-kitchen/internal/repository"
	"avito-kitchen/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://kitchen_user:kitchen_password@localhost:5432/avito_kitchen?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}

	repo := repository.NewRepository(pool)
	kitchenSvc := service.NewKitchenService(repo)
	handler := apihttp.NewHandler(kitchenSvc)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler.InitRoutes(),
	}

	go func() {
		log.Println("Starting Kitchen API on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down API...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("API stopped gracefully")
}
