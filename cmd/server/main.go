package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/choonhong/user-analytics/internal/api"
	"github.com/choonhong/user-analytics/internal/database"
	"github.com/choonhong/user-analytics/internal/repository"
	"github.com/choonhong/user-analytics/internal/service"
)

func main() {
	addr := envOr("ADDR", ":8080")

	ctx := context.Background()
	client, db, err := database.Open(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer client.Close() //nolint:errcheck
	defer db.Close()     //nolint:errcheck

	repo := repository.NewLoginRepository(client)
	svc := service.NewAnalyticsService(repo)
	handler := api.NewHandler(svc)
	router := api.NewRouter(handler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s (postgres)", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
