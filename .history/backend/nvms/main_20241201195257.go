package main

import (
	"fmt"
	"net/http"
	"nvms/lib"
	"os"
	"strings"

	"aidanwoods.dev/go-paseto"
	spinhttp "github.com/fermyon/spin-go-sdk/http"
)
var (
    // Initialize during startup with key from environment/config
    serviceKey paseto.V4SymmetricKey
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router := initRouter()
		router.ServeHTTP(w, r)
	})
   
}
func initRouter() *spinhttp.Router{
	router := spinhttp.NewRouter()
	router.GET("/", handle)
	return router;
}
func handle
func main() {
    // Main function is required for the Go program to run
    // You can add any initialization code here if needed
}