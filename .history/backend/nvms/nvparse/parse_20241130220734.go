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
        
    }
	fmt.Println("Current working directory:", dir)
	dirs, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return lib.NVMS{}, err
	}
	for _, d := range dirs {
		fmt.Println(d.Name())
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
