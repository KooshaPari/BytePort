package projectManager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nvms/lib"
	"nvms/models"
)




func TerminateProject(w http.ResponseWriter, r *http.Request) {
 /*Get Project, User from Req -> Deployments from DeploymentsJSON, loop thru call a terminate resource func(analyze service type choose appropriate termination function)*/
 var project models.Project; var user models.User;
 project, _, err := readBody(w, r)
 if err != nil {
	 return
 }
 var deployments map[string]models.Instance
 err = json.Unmarshal([]byte(project.DeploymentsJSON), &deployments)
 if err != nil {
	 http.Error(w, "Error parsing JSON", http.StatusBadRequest)
	 return
 }

 fmt.Println("Terminating project:", project.Name)
// Get managers for Windows termination
dockerManager, err := lib.GetDockerManager()
if err != nil {
	fmt.Println("Error getting Docker manager:", err)
	// Continue with termination even if Docker manager fails
}

tunnelManager, err := lib.GetTunnelManager()
if err != nil {
	fmt.Println("Error getting tunnel manager:", err)
	// Continue with termination even if tunnel manager fails
}

storageManager, err := lib.GetStorageManager()
if err != nil {
	fmt.Println("Error getting storage manager:", err)
	// Continue with termination even if storage manager fails
}

// Stop and remove Docker containers
if dockerManager != nil {
	containers, err := dockerManager.ListProjectContainers(project.Name)
	if err != nil {
		fmt.Println("Error listing containers:", err)
	} else {
		for _, container := range containers {
			fmt.Println("Stopping container:", container.Name)
			err = dockerManager.StopContainer(container.ContainerID)
			if err != nil {
				fmt.Println("Error stopping container:", err)
			}

			fmt.Println("Removing container:", container.Name)
			err = dockerManager.RemoveContainer(container.ContainerID)
			if err != nil {
				fmt.Println("Error removing container:", err)
			}
		}
	}
}

// Stop and remove tunnel
if tunnelManager != nil {
	fmt.Println("Removing tunnel for project:", project.Name)
	err = tunnelManager.RemoveProjectTunnel(project.Name)
	if err != nil {
		fmt.Println("Error removing tunnel:", err)
	}
}

// Clean up project files
if storageManager != nil {
	fmt.Println("Removing project files for:", project.Name)
	err = storageManager.RemoveProject(project.Name)
	if err != nil {
		fmt.Println("Error removing project files:", err)
	}
}
w.WriteHeader(http.StatusOK)
w.Write([]byte(`{"message": "Project terminated successfully"}`))
fmt.Println("Project Terminated")
}