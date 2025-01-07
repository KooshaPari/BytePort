package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/models"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"gopkg.in/yaml.v2"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		// receive NVMS Config, Parse into Object and Validate Contents, Return OBJ or Error
		var nvms models.NVMS; var nvmsString string
		body, err := io.ReadAll(r.Body)
        if err != nil {
            fmt.Println("Error reading request body: ", err)
            http.Error(w, "Error reading request body", http.StatusInternalServerError)
            return
        }
        defer r.Body.Close()
        fmt.Println("Parsing JSON...")
        err = json.Unmarshal(body, &nvmsString)
        if err != nil {
            http.Error(w, "Error parsing JSON", http.StatusBadRequest)
            return
        }
		w.Header().Set("Content-Type", "text/plain")
		 
	})
}

func main() {}
 
func parseNVMSConfig(yamlContent string) (*models.NVMS, error) {
    fmt.Printf("Parsing YAML content: %s\n", yamlContent) // Debug log
    
	config := &models.NVMS{
		Services: []models.Service{},
	}
    
    // Validate YAML content
    if strings.TrimSpace(yamlContent) == "" {
        return nil, fmt.Errorf("empty YAML content")
    }

    // Parse YAML with error handling
    err := yaml.Unmarshal([]byte(yamlContent), config)
    if err != nil {
        return nil, fmt.Errorf("YAML parsing error: %w", err)
    }

    // Validate required fields
    if config.Name == "" {
        return nil, fmt.Errorf("missing required field: NAME")
    }
    if len(config.Services) == 0 {
        return nil, fmt.Errorf("no services defined in YAML")
    }

    // Validate each service
	found := false
    for name, svc := range config.Services {
        if svc.Path == "" {
            return nil, fmt.Errorf("service %s missing PATH", name)
        }
        if svc.Port == 0 {
            return nil, fmt.Errorf("service %s missing PORT", name)
        }
		if svc.Name == "main"{
			if(found){
				return nil, fmt.Errorf("service main already defined", name)
			}else{
				found = true
			}
		}
    }
	if (!found){
		return nil, fmt.Errorf("service main not defined")
	}

    return config, nil
}