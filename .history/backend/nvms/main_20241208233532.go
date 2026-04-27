package main

import (
	"net/http"
	"nvms/lib"

	"nvms/deploy"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
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
func validateAction(func()) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
        lib.AuthMiddleware(w, r)
        handler(w, r)
    }
}
func main() {}