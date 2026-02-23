package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ava-sales/internal/config"
	"ava-sales/internal/db"
	"ava-sales/internal/handlers"
	"ava-sales/internal/repository"
	"ava-sales/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	store := repository.NewStore(pool)
	ticketService := service.NewTicketService(store)
	agencyService := service.NewAgencyService(store)

	router := handlers.NewRouter(ticketService, agencyService, store, cfg.JWTSecret)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if serveErr := server.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Fatalf("server failed: %v", serveErr)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
