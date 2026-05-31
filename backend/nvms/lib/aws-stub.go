package lib

import (
	"errors"
	"nvms/models"
)

// AWS functions are disabled for Windows deployment
// These are stub functions to maintain compatibility

func GetAWSCredentials(user models.User) (string, string, error) {
	return "", "", errors.New("AWS deployment disabled - using Windows deployment")
}

func PushToS3(zipBall []byte, AccessKey string, SecretKey string, ProjectName string) (LocalStorageInfo, error) {
	// Redirect to local storage
	sm, err := GetStorageManager()
	if err != nil {
		return LocalStorageInfo{}, err
	}
	return sm.PushToLocalStorage(zipBall, ProjectName)
}

func DeployEC2(AccessKey string, SecretKey string, Bucket LocalStorageInfo, service models.Service, fileMap []string) ([]DockerInstanceInfo, error) {
	// Redirect to Docker deployment
	dm, err := GetDockerManager()
	if err != nil {
		return nil, err
	}
	
	container, err := dm.CreateAndStartContainer(service, Bucket.Path)
	if err != nil {
		return nil, err
	}
	
	return []DockerInstanceInfo{*container}, nil
}

func AwaitInitialization(accessKey, secretKey string, instanceIDs []string) error {
	// For Docker containers, they're ready when they start
	return nil
}

func ProvisionNetwork(accessKey, secretKey, projectName string) (interface{}, string, string, error) {
	// Redirect to tunnel management
	tm, err := GetTunnelManager()
	if err != nil {
		return nil, "", "", err
	}
	
	// This is a simplified version - in real deployment, services would be passed
	projectURL := "https://" + projectName + ".yourdomain.com"
	return nil, "local", projectURL, nil
}

func CreateALBListener(accessKey, secretKey, projectName, lbArn, vpcId, instanceID string, port int) (string, string, error) {
	// Not needed for Windows deployment
	return "", "", nil
}

func RegisterService(accessKey, secretKey, lbArn, projectName, serviceName, vpcId, instanceID string, port int) (string, error) {
	// Not needed for Windows deployment
	return "", nil
}

func SetListenerRules(accessKey, secretKey, listenArn, tgArn, serviceName string, priority int) error {
	// Not needed for Windows deployment
	return nil
}

func TerminateS3(resource models.AWSResource, accessKey, secretKey string) error {
	// Redirect to local storage cleanup
	sm, err := GetStorageManager()
	if err != nil {
		return err
	}
	
	// Extract project name from resource ID
	projectName := resource.ID
	return sm.RemoveProject(projectName)
}

func TerminateEC2(resource models.AWSResource, accessKey, secretKey string) error {
	// Redirect to Docker container termination
	dm, err := GetDockerManager()
	if err != nil {
		return err
	}
	
	// Resource ID should be container ID
	err = dm.StopContainer(resource.ID)
	if err != nil {
		return err
	}
	
	return dm.RemoveContainer(resource.ID)
}

func TerminateALB(resource models.AWSResource, accessKey, secretKey string) error {
	// Not needed for Windows deployment
	return nil
}

func TerminateTargetGroup(resource models.AWSResource, accessKey, secretKey string) error {
	// Not needed for Windows deployment
	return nil
}
