package nvparse;

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"nvms/lib"
)

func Parse(file string) (lib.NVMS, error) {
	fmt.Println("Parsing file", file)
	// check if file exists
	let valid, err := os.ReadFile(file)
}
