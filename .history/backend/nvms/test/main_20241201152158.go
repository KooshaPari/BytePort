package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")

fmt.Println("out:", outb.String(), "err:", errb.String())
		var file = "byteport.nvms"
	if _, err := os.Stat(file); os.IsNotExist(err) {
            fmt.Println("File does not exist:", file);
			fmt.Println("Error: ", err);
        }
		 file = "/byteport.nvms"
	if _, err := os.Stat(file); os.IsNotExist(err) {
            fmt.Println("File does not exist:", file);
			fmt.Println("Error: ", err);
        }
		 file = "/test/byteport.nvms"
	if _, err := os.Stat(file); os.IsNotExist(err) {
            fmt.Println("File does not exist:", file);
			fmt.Println("Error: ", err);
        }
	})
	
	
		

}

func main() {}
