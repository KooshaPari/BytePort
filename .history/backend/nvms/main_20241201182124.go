package main

import (
	"fmt"
	"net/http"

	"aidanwoods.dev/go-paseto"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)
var(
	serviceKey paseto.V4SymmetricKey
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// router
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintln(w, "Hello Fermyon3!")
		case http.MethodPost:
			fmt.Fprintln(w, "Hello Fermyon2!")
		}
		
	})
}

func main() {}
