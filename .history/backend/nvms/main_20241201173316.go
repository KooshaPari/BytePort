package main

import (
	"fmt"
	"net/http"


	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// router
		
		fmt.Fprintln(w, "Parsed file", file, ":", nvms)
	})
}

func main() {}
