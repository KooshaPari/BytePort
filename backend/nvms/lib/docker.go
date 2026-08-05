// lib/docker.go - compatibility entry point for BytePort container lifecycle.
//
// The historical DockerManager API is retained for callers, but Docker is no
// longer a supported runtime.  Commands are rendered for the selected
// Podman/WSL Containers/Apple Containers adapter and are never executed here.
package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"nvms/models"
)

// ContainerRuntime identifies a supported local OCI-compatible backend.
type ContainerRuntime string

const (
	RuntimePodman          ContainerRuntime = "podman"
	RuntimeWSLContainers   ContainerRuntime = "wsl-containers"
	RuntimeAppleContainers ContainerRuntime = "apple-containers"
)

// ContainerManager owns lifecycle command generation for a selected backend.
// It deliberately does not invoke the backend; the generated script is the
// boundary consumed by the host adapter.
type ContainerManager struct {
	networkName    string
	runtime        ContainerRuntime
	runtimeCommand string
	mutex          sync.RWMutex
}

// DockerManager is retained as a source-compatible alias. New callers should
// use ContainerManager and the explicit runtime constructors instead.
type DockerManager = ContainerManager

// ParseContainerRuntime validates the configured adapter. Docker values are
// rejected explicitly so a stale configuration cannot execute Docker.
func ParseContainerRuntime(value string) (ContainerRuntime, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(RuntimePodman):
		return RuntimePodman, nil
	case string(RuntimeWSLContainers), "wslc", "wsl-containers.exe":
		return RuntimeWSLContainers, nil
	case string(RuntimeAppleContainers), "apple", "container":
		return RuntimeAppleContainers, nil
	case "docker", "docker-desktop":
		return "", fmt.Errorf("Docker runtime is not supported; choose podman, wsl-containers, or apple-containers")
	default:
		return "", fmt.Errorf("unsupported container runtime %q; choose podman, wsl-containers, or apple-containers", value)
	}
}

func runtimeCommand(runtime ContainerRuntime) string {
	switch runtime {
	case RuntimePodman:
		return "podman"
	case RuntimeWSLContainers:
		return "wslc"
	case RuntimeAppleContainers:
		return "container"
	default:
		return ""
	}
}

func configuredContainerRuntime() (ContainerRuntime, error) {
	value := os.Getenv("BYTEPORT_CONTAINER_RUNTIME")
	if value == "" {
		value = os.Getenv("CONTAINER_RUNTIME")
	}
	return ParseContainerRuntime(value)
}

func newContainerManager(runtime ContainerRuntime) (*ContainerManager, error) {
	command := runtimeCommand(runtime)
	if command == "" {
		return nil, fmt.Errorf("unsupported container runtime %q", runtime)
	}
	return &ContainerManager{
		networkName:    "byteport-network",
		runtime:        runtime,
		runtimeCommand: command,
	}, nil
}

// NewContainerManager selects BYTEPORT_CONTAINER_RUNTIME (or
// CONTAINER_RUNTIME) and defaults to Podman. It performs no runtime probing or
// process execution; call ValidateRuntime when the host is ready to launch.
func NewContainerManager() (*ContainerManager, error) {
	runtime, err := configuredContainerRuntime()
	if err != nil {
		return nil, err
	}
	return newContainerManager(runtime)
}

// NewContainerManagerForRuntime creates a deterministic adapter without
// probing the host. This is useful for capability planning and static tests.
func NewContainerManagerForRuntime(runtime ContainerRuntime) (*ContainerManager, error) {
	return newContainerManager(runtime)
}

// ValidateRuntime checks whether the selected adapter executable is available
// without starting it. A missing backend is reported instead of being treated
// as a working Docker-compatible runtime.
func (dm *ContainerManager) ValidateRuntime() error {
	if _, err := exec.LookPath(dm.runtimeCommand); err != nil {
		return fmt.Errorf("container runtime %q (%s) is unavailable: %w", dm.runtime, dm.runtimeCommand, err)
	}
	return nil
}

// Runtime returns the selected adapter name.
func (dm *ContainerManager) Runtime() ContainerRuntime {
	return dm.runtime
}

// RuntimeCommand returns the executable name used by generated scripts.
func (dm *ContainerManager) RuntimeCommand() string {
	return dm.runtimeCommand
}

type DockerInstanceInfo struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	Port        int    `json:"port"`
	Status      string `json:"status"`
	ProjectName string `json:"project_name"`
	ServiceName string `json:"service_name"`
	ImageTag    string `json:"image_tag"`
	InstanceID  string `json:"instance_id"`
	Region      string `json:"region"`
}

var dockerManagerInstance *DockerManager
var dockerManagerOnce sync.Once

func GetDockerManager() (*DockerManager, error) {
	var err error
	dockerManagerOnce.Do(func() {
		dockerManagerInstance, err = NewContainerManager()
	})
	return dockerManagerInstance, err
}

func NewDockerManager() (*DockerManager, error) {
	return NewContainerManager()
}

func (dm *DockerManager) ensureNetwork() error {
	return nil
}

func (dm *DockerManager) CreateAndStartContainer(service models.Service, projectPath string) (*DockerInstanceInfo, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	imageTag := fmt.Sprintf("byteport-%s-%s:latest", service.ProjectName, service.Name)
	if err := dm.buildImage(projectPath, service.Path, imageTag, service); err != nil {
		return nil, fmt.Errorf("failed to prepare container image inputs: %w", err)
	}

	containerName := fmt.Sprintf("byteport-%s-%s", service.ProjectName, service.Name)
	return &DockerInstanceInfo{
		ContainerID: containerName,
		Name:        containerName,
		Port:        service.Port,
		Status:      "prepared",
		ProjectName: service.ProjectName,
		ServiceName: service.Name,
		ImageTag:    imageTag,
		InstanceID:  containerName,
		Region:      "local",
	}, nil
}

func (dm *DockerManager) removeExistingContainer(containerName string) {
}

func (dm *DockerManager) buildImage(projectPath, servicePath, imageTag string, service models.Service) error {
	fullServicePath := filepath.Join(projectPath, servicePath)
	containerfilePath := filepath.Join(fullServicePath, "Containerfile")
	dockerfilePath := filepath.Join(fullServicePath, "Dockerfile")
	if _, dockerfileErr := os.Stat(dockerfilePath); os.IsNotExist(dockerfileErr) {
		if _, containerfileErr := os.Stat(containerfilePath); !os.IsNotExist(containerfileErr) {
			return dm.writeContainerCommandFile(fullServicePath, imageTag, service)
		}
		dockerfile := dm.generateDockerfile(fullServicePath, service)
		if err := os.WriteFile(containerfilePath, []byte(dockerfile), 0644); err != nil {
			return fmt.Errorf("failed to create Containerfile: %w", err)
		}
	}

	return dm.writeContainerCommandFile(fullServicePath, imageTag, service)
}

func (dm *DockerManager) generateDockerfile(servicePath string, service models.Service) string {
	if dm.fileExists(filepath.Join(servicePath, "package.json")) {
		return dm.generateNodeDockerfile(service.Port)
	}
	if dm.fileExists(filepath.Join(servicePath, "go.mod")) {
		return dm.generateGoDockerfile(service.Port)
	}
	if dm.fileExists(filepath.Join(servicePath, "requirements.txt")) {
		return dm.generatePythonDockerfile(service.Port)
	}
	if dm.fileExists(filepath.Join(servicePath, "Cargo.toml")) {
		return dm.generateRustDockerfile(service.Port)
	}
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

func (dm *DockerManager) StopContainer(containerID string) error {
	return nil
}

func (dm *DockerManager) RemoveContainer(containerID string) error {
	return nil
}

func (dm *DockerManager) GetContainerStatus(containerID string) (string, error) {
	return "prepared", nil
}

func (dm *DockerManager) ListProjectContainers(projectName string) ([]DockerInstanceInfo, error) {
	return []DockerInstanceInfo{}, nil
}

func (dm *DockerManager) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func (dm *DockerManager) Close() error {
	return nil
}

func (dm *DockerManager) writeContainerCommandFile(servicePath, imageTag string, service models.Service) error {
	containerName := fmt.Sprintf("byteport-%s-%s", service.ProjectName, service.Name)
	commandText := fmt.Sprintf(`%[5]s network create %[1]s 2>$null
%[5]s build --tag %[2]s .
%[5]s rm --force %[3]s 2>$null
%[5]s run --detach --name %[3]s --network %[1]s --restart unless-stopped --publish %[4]d:%[4]d --workdir /app %[2]s
`, dm.networkName, imageTag, containerName, service.Port, dm.runtimeCommand)

	commandPath := filepath.Join(servicePath, "byteport-container.ps1")
	return os.WriteFile(commandPath, []byte(commandText), 0644)
}

// writeDockerCommandFile remains as an internal compatibility wrapper for
// older package tests; it writes the neutral adapter script and never emits a
// Docker command.
func (dm *DockerManager) writeDockerCommandFile(servicePath, imageTag string, service models.Service) error {
	return dm.writeContainerCommandFile(servicePath, imageTag, service)
}
