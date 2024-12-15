package main

import (
	"fmt"
	"net/http"

	"nvms/deploy"
	"nvms/lib"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
	"github.com/julienschmidt/httprouter"
)


func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router := initRouter()
		router.ServeHTTP(w, r)
	})
   
}
func initRouter() *spinhttp.Router{
	router := spinhttp.NewRouter()
	router.POST("/",  validateAction(deploy.DeployProject))
	router.POST("/deploy",  validateAction(deploy.DeployProject))
	return router;
}
func validateAction(handler http.HandlerFunc) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
        lib.AuthMiddleware(w, r)
		
        handler(w, r)
    }
}
func main() {}