package nvparse

import (
	"fmt"
	"nvms/lib"
	"os"

	"gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"
)

func Parse(file string) (lib.NVMS, error) {
	fmt.Println("Parsing file", file);
	// check if file exists
	valid, err := os.ReadFile(file)
	if err != nil {
		return lib.NVMS{}, err
	}
	fmt.Println("Found file: ", file);
	result, err := yaml.Unmarshal(valid, lib.NVMS{})
	if err != nil {
		return lib.NVMS{}, err
	}
	return result.(lib.NVMS), nil
}
