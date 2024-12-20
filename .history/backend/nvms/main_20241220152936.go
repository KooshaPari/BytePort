package main

import (
	"net/http"
	Deploy "nvms/deploy"

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
	router.POST("/",  validateAction(Deploy.DeployProject))
	router.POST("/deploy",  validateAction(Deploy.DeployProject))
	return router;
}
func validateAction(handler http.HandlerFunc) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
        /*err := lib.AuthMiddleware(w, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {*/
        handler(w, r)
		/*}
		return*/
    }
}
func main() {}