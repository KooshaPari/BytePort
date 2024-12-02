package nvparse

import (
	"fmt"
	"nvms/lib"
	"os"
)

func Parse(file string) (lib.NVMS, error) {
	fmt.Println("Parsing file", file);
	// check if file exists
	valid, err := os.ReadFile(file)
}
