package main

import (
	"net/http"
	"nvms/models"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		var project models.Project;
		body, err := ioutil.ReadAll(r.Body)
		
	})
}

func main() {}
