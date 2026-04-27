package main

import (
	"fmt"
	"net/http"
	"nvms/nvparse"
	"nvms/lib"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")
		var file = "examples/byteport.nvms"
		
	})
}

func main() {}
