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
	router.GET("/",  validateAction(testhandler))
	return router;
}
func validateAction(handler http.HandlerFunc) http.Handle {
    return func(w http.ResponseWriter, r *http.Request) {
        lib.AuthMiddleware(w, r)
        handler(w, r)
    }
}
func testhandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
}
func main() {
    // Main function is required for the Go program to run
/*************  ✨ Codeium Command ⭐  *************/
// main is the entry point for the application.
/******  20286aae-38d1-45ec-9900-9d9be89d1d00  *******/    // You can add any initialization code here if needed
}