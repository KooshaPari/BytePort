package lib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nvms/models"
)

func TestDefaultRuntimeIsPodman(t *testing.T) {
	t.Setenv("BYTEPORT_CONTAINER_RUNTIME", "")
	t.Setenv("CONTAINER_RUNTIME", "")

	manager, err := NewContainerManager()
	if err != nil {
		t.Fatalf("NewContainerManager() error = %v", err)
	}
	if manager.Runtime() != RuntimePodman {
		t.Fatalf("default runtime = %q, want %q", manager.Runtime(), RuntimePodman)
	}
	if manager.RuntimeCommand() != "podman" {
		t.Fatalf("default runtime command = %q, want podman", manager.RuntimeCommand())
	}
}

func TestDockerRuntimeIsRejected(t *testing.T) {
	for _, value := range []string{"docker", "docker-desktop", "DOCKER"} {
		t.Run(value, func(t *testing.T) {
			if _, err := ParseContainerRuntime(value); err == nil {
				t.Fatalf("ParseContainerRuntime(%q) unexpectedly succeeded", value)
			} else if !strings.Contains(strings.ToLower(err.Error()), "not supported") {
				t.Fatalf("ParseContainerRuntime(%q) error = %v", value, err)
			}
		})
	}
}

func TestContainerCommandRenderingUsesSelectedAdapter(t *testing.T) {
	service := models.Service{ProjectName: "demo", Name: "api", Port: 8080}
	for _, test := range []struct {
		name    string
		runtime ContainerRuntime
		command string
	}{
		{name: "podman", runtime: RuntimePodman, command: "podman"},
		{name: "wsl-containers", runtime: RuntimeWSLContainers, command: "wslc"},
		{name: "apple-containers", runtime: RuntimeAppleContainers, command: "container"},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager, err := NewContainerManagerForRuntime(test.runtime)
			if err != nil {
				t.Fatalf("NewContainerManagerForRuntime() error = %v", err)
			}
			dir := t.TempDir()
			if err := manager.writeContainerCommandFile(dir, "byteport-demo-api:latest", service); err != nil {
				t.Fatalf("writeContainerCommandFile() error = %v", err)
			}
			commandPath := filepath.Join(dir, "byteport-container.ps1")
			contents, err := os.ReadFile(commandPath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", commandPath, err)
			}
			commandText := string(contents)
			if !strings.Contains(commandText, test.command+" network create") {
				t.Fatalf("command does not select %q:\n%s", test.command, commandText)
			}
			if strings.Contains(strings.ToLower(commandText), "docker") {
				t.Fatalf("generated command contains Docker token:\n%s", commandText)
			}
			if _, err := os.Stat(filepath.Join(dir, "byteport-docker.ps1")); !os.IsNotExist(err) {
				t.Fatalf("legacy Docker command file unexpectedly exists (err=%v)", err)
			}
		})
	}
}

func TestBuildImageWritesNeutralContainerfile(t *testing.T) {
	manager, err := NewContainerManagerForRuntime(RuntimePodman)
	if err != nil {
		t.Fatalf("NewContainerManagerForRuntime() error = %v", err)
	}
	project := t.TempDir()
	servicePath := filepath.Join(project, "api")
	if err := os.MkdirAll(servicePath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(servicePath, "go.mod"), []byte("module example.test/api\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}
	service := models.Service{ProjectName: "demo", Name: "api", Path: "api", Port: 8080}
	if err := manager.buildImage(project, service.Path, "byteport-demo-api:latest", service); err != nil {
		t.Fatalf("buildImage() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(servicePath, "Containerfile")); err != nil {
		t.Fatalf("generated Containerfile missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(servicePath, "byteport-container.ps1")); err != nil {
		t.Fatalf("generated adapter script missing: %v", err)
	}
}
