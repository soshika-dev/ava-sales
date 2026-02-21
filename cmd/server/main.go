package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"ava-sales/internal/config"
	"ava-sales/internal/db"
	"ava-sales/internal/handlers"
	"ava-sales/internal/repositories"
	"ava-sales/internal/router"
	"ava-sales/internal/services"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer pool.Close()

	agencyRepo := repositories.NewAgencyRepository(pool)
	ticketRepo := repositories.NewTicketRepository(pool)
	ticketService := services.NewTicketService(ticketRepo)

	r := router.New(
		handlers.NewAgencyHandler(agencyRepo),
		handlers.NewTicketHandler(ticketRepo, ticketService),
	)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
