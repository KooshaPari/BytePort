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
result := lib.NVMS{};
	if _, err := os.Stat(file); os.IsNotExist(err) {
            fmt.Println("File does not exist:", file);
			fmt.Println("Error: ", err);
            return result, nil
        }
	valid, err := os.ReadFile(file)
	if err != nil {
		return lib.NVMS{}, err
	}
	fmt.Println("Found file: ", file);
	
	err = yaml.Unmarshal(valid, result)
	if err != nil {
		return lib.NVMS{}, err
	}
	return result, nil
}
