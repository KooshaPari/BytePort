package lib

import (
	"fmt"
	"nvms/models"
	"path/filepath"
	"strings"
)

// S3DeploymentInfo contains information about deployed S3 resources
type S3DeploymentInfo struct {
    BucketName   string   // Name of the created bucket
    ObjectKey    string   // Key of the uploaded object (e.g., "src.zip")
    Region       string   // AWS region where bucket was created
    BucketARN    string   // ARN of the bucket
    ObjectURL    string   // Full URL to access the object
    ContentHash  string   // SHA256 hash of uploaded content
}

// EC2InstanceInfo contains information about deployed EC2 instances
type EC2InstanceInfo struct {
    InstanceID       string   // EC2 instance ID
    PublicIP         string   // Public IP address
    PrivateIP        string   // Private IP address
    PublicDNS        string   // Public DNS name
    PrivateDNS       string   // Private DNS name
    State            string   // Current instance state
    KeyPairName      string   // Name of the SSH key pair
    SecurityGroups   []string // List of security group IDs
    SubnetID         string   // Subnet ID where instance is launched
    Region           string   // AWS region where instance is launched
}
type BuilderReq struct {
	ZipBall     []byte `json:"zipball"`
	AccessKey   string `json:"accessKey"`
	SecretKey   string `json:"secretKey"`
	ProjectName string `json:"projectName"`
}
func getServiceEndpoint(service string) string {
    if AWSEndpointBase == "http://localhost.localstack.cloud:4566" {
        // LocalStack uses a single endpoint for all services
        return AWSEndpointBase
    }
    // AWS uses service-specific endpoints
    return fmt.Sprintf(AWSEndpointBase, service)
}


func GetAWSCredentials(user models.User)(string,string,error){
	eAccKey := user.AwsCreds.AccessKeyID
	eSecKey := user.AwsCreds.SecretAccessKey

	accesskey, err :=  DecryptSecret(eAccKey)
	if err != nil {
	 
		return "","",fmt.Errorf("Error decrypting access key")
	}
	secretkey, err :=  DecryptSecret(eSecKey)
	if err != nil {
		  
		return "","",fmt.Errorf("Error decrypting secret key")
	}
    return accesskey,secretkey,nil

}
 