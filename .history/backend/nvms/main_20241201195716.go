package main

import (
	"net/http"
	"nvms/lib"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)


func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router := initRouter()
		router.ServeHTTP(w, r)
	})
   
}
func initRouter() *spinhttp.Router {
    router := spinhttp.NewRouter()
    router.Use(lib.AuthMiddleware) // apply middleware to all routes
    router.GET("/", func(w http.ResponseWriter, r *http.Request) {
        // handler for root route
    })
    // create a group for protected routes
    protected := router.Group("/protected")
    protected.Use(lib.AuthMiddleware) // apply middleware to protected routes
    protected.GET("/secret", func(w http.ResponseWriter, r *http.Request) {
        // handler for secret route
    })
    return router
}

func main() {
    // Main function is required for the Go program to run
/*************  ✨ Codeium Command ⭐  *************/
// main is the entry point for the application.
/******  20286aae-38d1-45ec-9900-9d9be89d1d00  *******/    // You can add any initialization code here if needed
}