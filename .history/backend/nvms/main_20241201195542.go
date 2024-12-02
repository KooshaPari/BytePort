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
func initRouter() *spinhttp.Router{
	router := spinhttp.NewRouter()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		lib.AuthMiddleware(w, r)
	})
	return router;
}

func main() {
    // Main function is required for the Go program to run
    // You can add any initialization code here if needed
}