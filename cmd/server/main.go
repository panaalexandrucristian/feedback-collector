package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/panaalexandrucristian/feedback-collector/internal/api"
	"github.com/panaalexandrucristian/feedback-collector/internal/config"
	"github.com/panaalexandrucristian/feedback-collector/internal/db"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting server in %s mode on port %d\n", cfg.Environment, cfg.Port)

	// Connect to database
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Setup router
	router := api.SetupRouter(cfg, database)

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.GetPortString(),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give the server 5 seconds to finish ongoing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
