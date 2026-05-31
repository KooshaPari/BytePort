// lib/docker.go - Docker management for BytePort Windows
package lib

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"nvms/models"
)

type DockerManager struct {
	client *client.Client
	ctx    context.Context
	mutex  sync.RWMutex
}

type DockerInstanceInfo struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	Port        int    `json:"port"`
	Status      string `json:"status"`
	ProjectName string `json:"project_name"`
	ServiceName string `json:"service_name"`
	ImageTag    string `json:"image_tag"`
	InstanceID  string `json:"instance_id"` // For compatibility with existing code
	Region      string `json:"region"`     // For compatibility with existing code
}

var dockerManagerInstance *DockerManager
var dockerManagerOnce sync.Once

func GetDockerManager() (*DockerManager, error) {
	var err error
	dockerManagerOnce.Do(func() {
		dockerManagerInstance, err = NewDockerManager()
	})
	return dockerManagerInstance, err
}

func NewDockerManager() (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	dm := &DockerManager{
		client: cli,
		ctx:    context.Background(),
	}

	// Ensure BytePort network exists
	err = dm.ensureNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure network: %w", err)
	}

	return dm, nil
}

func (dm *DockerManager) ensureNetwork() error {
	networkName := "byteport-network"
	
	// Check if network exists
	networks, err := dm.client.NetworkList(dm.ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, net := range networks {
		if net.Name == networkName {
			return nil // Network already exists
		}
	}

	// Create network
	_, err = dm.client.NetworkCreate(dm.ctx, networkName, types.NetworkCreate{
		Driver: "bridge",
	})
	return err
}

func (dm *DockerManager) CreateAndStartContainer(service models.Service, projectPath string) (*DockerInstanceInfo, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Generate unique image tag
	imageTag := fmt.Sprintf("byteport-%s-%s:latest", service.ProjectName, service.Name)
	
	// Build Docker image
	err := dm.buildImage(projectPath, service.Path, imageTag, service)
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image: imageTag,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", service.Port)): {},
		},
		Env:        dm.buildEnvVars(service.Env),
		WorkingDir: "/app",
	}

	// Host configuration with port binding
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", service.Port)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", service.Port),
				},
			},
		},
		NetworkMode: "byteport-network",
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Network configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"byteport-network": {},
		},
	}

	// Container name
	containerName := fmt.Sprintf("byteport-%s-%s", service.ProjectName, service.Name)

	// Remove existing container if it exists
	dm.removeExistingContainer(containerName)

	// Create container
	resp, err := dm.client.ContainerCreate(
		dm.ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	err = dm.client.ContainerStart(dm.ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return &DockerInstanceInfo{
		ContainerID: resp.ID,
		Name:        containerName,
		Port:        service.Port,
		Status:      "running",
		ProjectName: service.ProjectName,
		ServiceName: service.Name,
		ImageTag:    imageTag,
		InstanceID:  resp.ID, // For compatibility
		Region:      "local", // For compatibility
	}, nil
}

func (dm *DockerManager) removeExistingContainer(containerName string) {
	// Try to stop and remove existing container
	containers, err := dm.client.ContainerList(dm.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				dm.client.ContainerStop(dm.ctx, container.ID, nil)
				dm.client.ContainerRemove(dm.ctx, container.ID, types.ContainerRemoveOptions{Force: true})
				return
			}
		}
	}
}

func (dm *DockerManager) buildImage(projectPath, servicePath, imageTag string, service models.Service) error {
	// Full path to service directory
	fullServicePath := filepath.Join(projectPath, servicePath)
	
	// Generate Dockerfile if it doesn't exist
	dockerfilePath := filepath.Join(fullServicePath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		dockerfile := dm.generateDockerfile(fullServicePath, service)
		err = os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
		if err != nil {
			return fmt.Errorf("failed to create Dockerfile: %w", err)
		}
	}

	// Create tar archive for build context
	buildContext, err := dm.createBuildContext(fullServicePath)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	// Build image
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
	}

	buildResponse, err := dm.client.ImageBuild(dm.ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer buildResponse.Body.Close()

	// Read build output (for logging)
	_, err = io.Copy(io.Discard, buildResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read build output: %w", err)
	}

	return nil
}

func (dm *DockerManager) generateDockerfile(servicePath string, service models.Service) string {
	// Detect project type and generate appropriate Dockerfile
	if dm.fileExists(filepath.Join(servicePath, "package.json")) {
		return dm.generateNodeDockerfile(service.Port)
	} else if dm.fileExists(filepath.Join(servicePath, "go.mod")) {
		return dm.generateGoDockerfile(service.Port)
	} else if dm.fileExists(filepath.Join(servicePath, "requirements.txt")) {
		return dm.generatePythonDockerfile(service.Port)
	} else if dm.fileExists(filepath.Join(servicePath, "Cargo.toml")) {
		return dm.generateRustDockerfile(service.Port)
	}
	
	// Default to Node.js
	return dm.generateNodeDockerfile(service.Port)
}

func (dm *DockerManager) generateNodeDockerfile(port int) string {
	return fmt.Sprintf(`FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE %d
CMD ["npm", "start"]`, port)
}

func (dm *DockerManager) generateGoDockerfile(port int) string {
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
CMD ["./main"]`, port)
}

func (dm *DockerManager) generatePythonDockerfile(port int) string {
	return fmt.Sprintf(`FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE %d
CMD ["python", "app.py"]`, port)
}

func (dm *DockerManager) generateRustDockerfile(port int) string {
	return fmt.Sprintf(`FROM rust:1.70 AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/target/release/app .
EXPOSE %d
CMD ["./app"]`, port)
}

func (dm *DockerManager) createBuildContext(contextPath string) (io.ReadCloser, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	err := filepath.Walk(contextPath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git and node_modules directories
		if strings.Contains(file, ".git") || strings.Contains(file, "node_modules") {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(contextPath, file)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return io.NopCloser(&buf), nil
}

func (dm *DockerManager) StopContainer(containerID string) error {
	timeout := time.Second * 30
	return dm.client.ContainerStop(dm.ctx, containerID, &timeout)
}

func (dm *DockerManager) RemoveContainer(containerID string) error {
	return dm.client.ContainerRemove(dm.ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (dm *DockerManager) GetContainerStatus(containerID string) (string, error) {
	containerJSON, err := dm.client.ContainerInspect(dm.ctx, containerID)
	if err != nil {
		return "", err
	}
	return containerJSON.State.Status, nil
}

func (dm *DockerManager) ListProjectContainers(projectName string) ([]DockerInstanceInfo, error) {
	containers, err := dm.client.ContainerList(dm.ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	var projectContainers []DockerInstanceInfo
	prefix := fmt.Sprintf("byteport-%s-", projectName)

	for _, container := range containers {
		for _, name := range container.Names {
			if strings.HasPrefix(strings.TrimPrefix(name, "/"), prefix) {
				serviceName := strings.TrimPrefix(strings.TrimPrefix(name, "/"), prefix)
				
				var port int
				for containerPort := range container.Ports {
					if containerPort.Type == "tcp" {
						port = int(containerPort.PrivatePort)
						break
					}
				}

				projectContainers = append(projectContainers, DockerInstanceInfo{
					ContainerID: container.ID,
					Name:        strings.TrimPrefix(name, "/"),
					Port:        port,
					Status:      container.Status,
					ProjectName: projectName,
					ServiceName: serviceName,
					InstanceID:  container.ID,
					Region:      "local",
				})
			}
		}
	}

	return projectContainers, nil
}

func (dm *DockerManager) buildEnvVars(envMap map[string]string) []string {
	var envVars []string
	for key, value := range envMap {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}
	return envVars
}

func (dm *DockerManager) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func (dm *DockerManager) Close() error {
	return dm.client.Close()
}
