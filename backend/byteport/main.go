package main

import (
	"byteport/lib"
	"byteport/models"
	"byteport/routes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/gorm"
)

// initTracer sets up the OTel tracer provider with a ConsoleSpanExporter.
func initTracer() (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create stdouttrace exporter: %w", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(routes.MetricsMiddleware())

	// Origins without an http/https scheme cannot be listed in AllowOrigins:
	// gin-contrib/cors panics on them ("bad origin: origins must contain '*' or
	// include http://,https://"). The macOS/iOS Tauri webview serves the app
	// from tauri://localhost, so it is admitted through AllowOriginFunc instead.
	// Omitting it made every preflight from the packaged desktop app fail with
	// 403, so the app could not reach this backend at all.
	httpOrigins := []string{
		"http://localhost:5173",
		"http://0.0.0.0:5173",
		// Windows/Linux Tauri webview origin.
		"http://tauri.localhost",
		"http://tauri.0.0.0.0:5173",
		"http://localhost:8081",
		"http://0.0.0.0:8081",
		"http://10.0.2.2:5173",
		"http://10.0.2.2:8081",
		// Add other needed origins
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins: httpOrigins,
		AllowOriginFunc: func(origin string) bool {
			if origin == "tauri://localhost" {
				return true
			}
			for _, allowed := range httpOrigins {
				if allowed == origin {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWildcard:    true, // Enable wildcard matching
		MaxAge:           12 * time.Hour,
	}))
	protected := r.Group("/")
	protected.Use(lib.AuthMiddleware())
	{

		protected.GET("/link", routes.LinkHandler)
		protected.POST("/link", routes.ValidateLink)
		protected.GET("/authenticate", routes.Authenticate)
		protected.GET("/instances", routes.GetInstances)
		protected.GET("/projects", routes.GetProjects)
		protected.GET("/api/github/repositories", routes.RetrieveRepositories)
		protected.POST("/deploy", routes.DeployProject)
		protected.POST("/terminate", routes.TerminateInstance)
		protected.GET("/user/:id/creds", routes.UpdateLink)
		protected.PUT("/user/:id/creds", routes.UpdateUser)
		//protected.GET("/github/status", routes.GitHubStatusHandler)
	}
	r.POST("/login", routes.Login)
	r.POST("/signup", routes.Signup)
	r.GET("/api/github/callback", routes.HandleCallback)
	r.GET("/health", routes.HealthHandler)
	r.GET("/metrics", routes.MetricsHandler)

	// gh webhook at /api/github/auth/webhook

	return r
}

// canonicalPort is the BytePort API port. It is the single default that every
// consumer targets: SvelteKit (src/lib/config.ts, getBaseUrl), the Tauri shell
// and its CSP (src-tauri/src/lib.rs, tauri.conf.json), .air.toml proxy_port,
// scripts/verify-frontend-boot.py and setup-windows.ps1.
//
// Override at runtime with PORT (documented in INSTALL.md/DEPLOYMENT.md) or
// BYTEPORT_API_PORT (written by setup-windows.ps1). PORT wins.
const canonicalPort = "8081"

// resolvePort returns the port to listen on, preferring PORT over
// BYTEPORT_API_PORT, and falling back to canonicalPort.
func resolvePort() string {
	for _, key := range []string{"PORT", "BYTEPORT_API_PORT"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return canonicalPort
}

func main() {
	ctx := context.Background()

	tp, err := initTracer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing tracer: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down tracer: %v\n", err)
		}
	}()

	err = lib.InitializeEncryptionKey()
	if err != nil {
		fmt.Printf("Error initializing encryption key: %v\n", err)
		os.Exit(1)
	}

	models.ConnectDatabase()
	err = lib.InitAuthSystem()
	if err != nil {
		fmt.Printf("Error initializing auth system: %v\n", err)
		os.Exit(1)
	}
	var temp models.GitSecret
	result := models.DB.First(&temp) // Retrieve the first entry
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Println("No entries found in git_secrets table (empty DB; starting fresh).")
		} else {
			fmt.Printf("Error retrieving data from git_secrets table: %v\n", result.Error)
			os.Exit(1)
		}
	}
	//models.DB.Exec("Delete from projects")
	r := setupRouter()
	//models.DB.Exec("Delete from users")

	go lib.StartTokenRefreshJob()
	addr := "0.0.0.0:" + resolvePort()
	fmt.Printf("BytePort API Server listening on %s\n", addr)
	if err := r.Run(addr); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
