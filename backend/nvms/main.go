package main

import (
	"fmt"
	"log"
	"net/http"
	"nvms/projectManager"
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
	router.POST("/", corsWrapper(projectManager.DeployProject))
	router.POST("/deploy", corsWrapper(projectManager.DeployProject))
	router.POST("/terminate", corsWrapper(projectManager.TerminateProject))

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

func main() {
	router := initRouter()

	port := os.Getenv("BYTEPORT_NVMS_PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("🚀 NVMS Service starting on port %s...\n", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}