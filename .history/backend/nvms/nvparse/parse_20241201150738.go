package nvparse

import (
	"fmt"
	"nvms/lib"
	"os"

	"gopkg.in/yaml.v3"
)

func Parse(file string) (lib.NVMS, error) {
	fmt.Println("Parsing file", file);
	// check if file exists'

	if _, err := os.Stat(file); os.IsNotExist(err) {
            fmt.Println("File does not exist:", file)
            return result, nil
        }
	valid, err := os.ReadFile(file)
	if err != nil {
		return lib.NVMS{}, err
	}
	fmt.Println("Found file: ", file);
	result := lib.NVMS{};
	err = yaml.Unmarshal(valid, result)
	if err != nil {
		return lib.NVMS{}, err
	}
	return result, nil
}
