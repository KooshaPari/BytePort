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
	dir, err := os.Getwd()
    if err != nil {
        fmt.Println("Error getting current working directory:", err)
        return
    }
    fmt.Println("Current working directory:", dir)
	if _, err := os.Stat(file); os.IsNotExist(err) {
            fmt.Println("File does not exist:", file)
            
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