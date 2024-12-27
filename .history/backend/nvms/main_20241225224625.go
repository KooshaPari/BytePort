package main

import (
	"net/http"
	"nvms/projectManager"
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
	router.POST("/",  validateAction(projectManager.DeployProject))
	router.POST("/deploy",  validateAction(projectManager.DeployProject))
	router.POST("/terminate",  validateAction(deploy.DeployProject))
	return router;
}
func validateAction(handler http.HandlerFunc) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
         err := lib.AuthMiddleware(w, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {
        handler(w, r)
		}
		return
    }
}
func main() {}