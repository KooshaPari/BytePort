package nvparse;

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

func Parse(file string) (NVMS, error) {
	fmt.Println("Parsing file", file)
	// check if file exists
	let valid, err := os.ReadFile(file)
}