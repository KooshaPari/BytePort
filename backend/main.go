package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/byteport/api/internal/container"
	"github.com/byteport/api/internal/infrastructure/otel"
	"github.com/byteport/api/lib"
	"github.com/byteport/api/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load orchestrator port before .env overrides
	orchestratorPort := os.Getenv("PORT")

	// Load environment configuration from backend/.env if present
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: .env file not found at %s, using environment variables", envPath)
	} else {
		log.Printf("Loaded environment from %s", envPath)
	}

	if orchestratorPort != "" {
		os.Setenv("PORT", orchestratorPort)
	}

	if err := lib.InitializeEncryptionKey(); err != nil {
		log.Fatalf("failed to initialise encryption key: %v", err)
	}

	models.ConnectDatabase()

	shutdownOTel := otel.InitOpenTelemetry()
	defer shutdownOTel()

	if err := lib.InitAuthSystem(); err != nil {
		log.Fatalf("failed to initialise auth system: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize dependency injection container
	containerInst := container.NewContainer(models.DB)
	log.Printf("✅ Dependency injection container initialized")

	server := NewAPIServer(containerInst)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀 BytePort API Server starting on %s", addr)
	log.Printf("📊 Environment: %s", env)
	log.Printf("🌐 API Documentation: http://localhost:%s/api/v1/docs", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: server.router.Handler(),
	}

	// Start server in a goroutine so we can listen for signals
	go func() {
		log.Printf("BytePort API Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal and gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("server stopped")
}
