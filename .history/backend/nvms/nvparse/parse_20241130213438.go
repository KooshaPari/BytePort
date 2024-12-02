package nvparse;

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

func Parse(file string) {
	fmt.Println("Parsing file", file)
	// check if file exists
	let valid, Error = os.ReadFile(file)
}
