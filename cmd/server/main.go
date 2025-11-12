package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devzen-shownote-generator/internal/config"
	"devzen-shownote-generator/internal/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create handler
	h := handlers.New(cfg)

	// Set up HTTP routes
	http.HandleFunc("/generate", h.HandleGenerate)
	http.HandleFunc("/trellohook", h.HandleTrelloHook)

	// Create server
	server := &http.Server{
		Addr:         "0.0.0.0:" + cfg.Port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server online at %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
