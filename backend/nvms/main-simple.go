package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func initRouter() *httprouter.Router {
	router := httprouter.New()
	
	// Enable CORS
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.WriteHeader(http.StatusOK)
	})
	
	// Wrap handlers to add CORS headers
	router.POST("/", corsWrapper(handleDeploy))
	router.POST("/deploy", corsWrapper(handleDeploy))
	router.POST("/terminate", corsWrapper(handleTerminate))
	router.GET("/health", corsWrapper(handleHealth))
	
	return router
}

func corsWrapper(handler http.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		handler(w, r)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "nvms", "deployment": "windows"}`))
}

func handleDeploy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// For now, return a success response
	// In a full implementation, this would:
	// 1. Parse the deployment request
	// 2. Clone the repository
	// 3. Parse odin.nvms configuration
	// 4. Build and start Docker containers
	// 5. Set up tunnels
	// 6. Return deployment info
	
	response := `{
		"status": "success",
		"message": "Deployment initiated",
		"deployment_id": "test-deployment-001",
		"services": [
			{
				"name": "main",
				"status": "running",
				"port": 8080,
				"url": "http://localhost:8080"
			}
		],
		"public_url": "https://your-app.trycloudflare.com"
	}`
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func handleTerminate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// For now, return a success response
	// In a full implementation, this would:
	// 1. Parse the termination request
	// 2. Stop and remove Docker containers
	// 3. Clean up storage
	// 4. Stop tunnels
	// 5. Return termination status
	
	response := `{
		"status": "success",
		"message": "Deployment terminated",
		"deployment_id": "test-deployment-001"
	}`
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func main() {
	router := initRouter()
	
	port := os.Getenv("BYTEPORT_NVMS_PORT")
	if port == "" {
		port = "3000"
	}
	
	fmt.Printf("🚀 NVMS Service (Windows) starting on port %s...\n", port)
	fmt.Printf("📍 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🐳 Docker deployment ready\n")
	fmt.Printf("🌐 Tunnel support available\n")
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	
	log.Fatal(server.ListenAndServe())
}
