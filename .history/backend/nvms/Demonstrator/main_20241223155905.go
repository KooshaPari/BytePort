package main

import (
	"fmt"
	"net/http"
	"nvms/models"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		var user models.User var project models.Project var nvms models.NVMS var readMe string
		var portfolioCreds models.Portfolio
		// receive Project Obj, Readme, NVMS 
		// pull Portfolio API Format
		// Generate Response (hand info above to ai langchain)
		// post to portfolio
		// return
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}

func main() {}
