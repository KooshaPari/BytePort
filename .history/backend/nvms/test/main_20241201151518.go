package main

import (
	"fmt"
	"net/http"
	"nvms/nvparse"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")
	})
	var file = "byteport.nvms"
		nvms, err := nvparse.Parse(file)
		if err != nil {
			fmt.Fprintln(w, "Error parsing file", file, ":", err)
			return
		}
		fmt.Fprintln(w, "Parsed file", file, ":", nvms)
}

func main() {}
