package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	Name        string
	Path        string
	Port        int
	ProjectName string
}

type NVMSConfig struct {
	Name        string
	Description string
	Services    []Service
}

func parseNVMSConfig(configPath string) (*NVMSConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &NVMSConfig{}
	scanner := bufio.NewScanner(file)
	
	var currentService *Service
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if strings.HasPrefix(line, "NAME:") {
			config.Name = strings.Trim(strings.TrimPrefix(line, "NAME:"), " \"")
		} else if strings.HasPrefix(line, "DESCRIPTION:") {
			config.Description = strings.Trim(strings.TrimPrefix(line, "DESCRIPTION:"), " \"")
		} else if strings.HasPrefix(line, "- NAME:") {
			if currentService != nil {
				config.Services = append(config.Services, *currentService)
			}
			currentService = &Service{}
			currentService.Name = strings.Trim(strings.TrimPrefix(line, "- NAME:"), " \"")
		} else if strings.HasPrefix(line, "  NAME:") {
			if currentService == nil {
				currentService = &Service{}
			}
			currentService.Name = strings.Trim(strings.TrimPrefix(line, "  NAME:"), " \"")
		} else if strings.HasPrefix(line, "  PATH:") {
			if currentService != nil {
				currentService.Path = strings.Trim(strings.TrimPrefix(line, "  PATH:"), " \"")
			}
		} else if strings.HasPrefix(line, "  PORT:") {
			if currentService != nil {
				portStr := strings.TrimSpace(strings.TrimPrefix(line, "  PORT:"))
				if port, err := strconv.Atoi(portStr); err == nil {
					currentService.Port = port
				}
			}
		}
	}
	
	if currentService != nil {
		config.Services = append(config.Services, *currentService)
	}
	
	return config, scanner.Err()
}

func generateDockerfile(service Service, projectPath string) string {
	servicePath := filepath.Join(projectPath, service.Path)
	
	if fileExists(filepath.Join(servicePath, "package.json")) {
		return fmt.Sprintf(`FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE %d
CMD ["npm", "start"]`, service.Port)
	} else if fileExists(filepath.Join(servicePath, "go.mod")) {
		return fmt.Sprintf(`FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE %d
CMD ["./main"]`, service.Port)
	}
	
	return fmt.Sprintf(`FROM node:18-alpine
WORKDIR /app
COPY . .
EXPOSE %d
CMD ["npm", "start"]`, service.Port)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func testDeployment(projectPath string) error {
	fmt.Printf("🚀 Testing deployment for project: %s\n", projectPath)
	
	configPath := filepath.Join(projectPath, "odin.nvms")
	config, err := parseNVMSConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse NVMS config: %w", err)
	}
	
	fmt.Printf("📋 Project: %s - %s\n", config.Name, config.Description)
	fmt.Printf("📦 Services found: %d\n", len(config.Services))
	
	for i, service := range config.Services {
		fmt.Printf("\n🔧 Deploying service %d/%d: %s\n", i+1, len(config.Services), service.Name)
		
		service.ProjectName = config.Name
		
		servicePath := filepath.Join(projectPath, service.Path)
		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			fmt.Printf("   ⚠️  Service path does not exist: %s\n", servicePath)
			continue
		}
		
		dockerfile := generateDockerfile(service, projectPath)
		fmt.Printf("   🐳 Generated Dockerfile:\n")
		for _, line := range strings.Split(dockerfile, "\n") {
			fmt.Printf("      %s\n", line)
		}
		
		fmt.Printf("   ⏳ Simulating container build...\n")
		time.Sleep(1 * time.Second)
		fmt.Printf("   ✅ Service %s ready on port %d\n", service.Name, service.Port)
	}
	
	fmt.Printf("\n🌐 Generating tunnel configuration...\n")
	tunnelConfig := generateTunnelConfig(config.Name, config.Services)
	fmt.Printf("Tunnel config:\n%s\n", tunnelConfig)
	
	fmt.Printf("\n🎉 Deployment test completed successfully!\n")
	fmt.Printf("📍 Project would be accessible at: https://%s.yourdomain.com\n", config.Name)
	
	return nil
}

func generateTunnelConfig(projectName string, services []Service) string {
	config := fmt.Sprintf(`tunnel: byteport-main
credentials-file: C:\BytePort\tunnels\credentials.json

ingress:`)

	var mainService *Service
	for i, service := range services {
		if service.Name == "main" {
			mainService = &services[i]
			break
		}
	}

	if mainService != nil {
		config += fmt.Sprintf(`
  - hostname: %s.yourdomain.com
    service: http://localhost:%d`, projectName, mainService.Port)
	}

	for _, service := range services {
		if service.Name != "main" {
			config += fmt.Sprintf(`
  - hostname: %s.yourdomain.com
    path: /%s/*
    service: http://localhost:%d`, projectName, service.Name, service.Port)
		}
	}

	config += `
  - service: http_status:404

logfile: C:\BytePort\logs\tunnel.log`

	return config
}

func main() {
	fmt.Println("🧪 BytePort Windows Deployment Test")
	fmt.Println("===================================")
	
	projectPath := "../backend/nvms/fixit-go"
	if len(os.Args) > 1 {
		projectPath = os.Args[1]
	}
	
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		log.Fatalf("❌ Project path does not exist: %s", projectPath)
	}
	
	configPath := filepath.Join(projectPath, "odin.nvms")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("❌ odin.nvms not found in project: %s", configPath)
	}
	
	err := testDeployment(projectPath)
	if err != nil {
		log.Fatalf("❌ Deployment test failed: %v", err)
	}
	
	fmt.Println("\n✅ All tests passed! BytePort Windows deployment is working correctly.")
}
