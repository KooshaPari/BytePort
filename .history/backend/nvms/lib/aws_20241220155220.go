package lib

import (
	"context"
	"encoding/base64"
	"fmt"

	aws "nvms/lib/awspin"
	ec2 "nvms/lib/awspin/ec2"
	awsnet "nvms/lib/awspin/network"
	r53 "nvms/lib/awspin/network/route53"
	"nvms/lib/awspin/s3"
	"nvms/models"
	"strings"

	"github.com/google/uuid"
)

func PushToS3(zipBall []byte, AccessKey string, SecretKey string, ProjectName string) (S3DeploymentInfo,error) {
	fmt.Println("Uploading to S3...")
 	cfg := aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://s3.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "s3",
	}
	ctx := context.Background()
	s3Client, err := s3.NewS3(cfg)
	if err != nil {
		fmt.Println(err)
		return S3DeploymentInfo{},err
	}
	fmt.Println("Created S3 client")
	bucketName := strings.ToLower(ProjectName) + "-bytebucket-" + uuid.New().String()
	err = s3Client.CreateBucket(ctx, bucketName)
	if err != nil {
			fmt.Println(err)
			return S3DeploymentInfo{},err
		}
		
	fmt.Println("Created bucket")
	err = s3Client.PutObject(ctx, bucketName, "src.zip", zipBall)
	if err != nil {
		fmt.Println(err)
		return S3DeploymentInfo{},err
	}
	fmt.Println("Uploaded to S3")
	// return uri/bucket name for later use
	info := S3DeploymentInfo{
        BucketName:  bucketName,
        ObjectKey:   "src.zip",
        Region:      "us-east-1",
        BucketARN:   fmt.Sprintf("arn:aws:s3:::%s", bucketName),
        ObjectURL:   fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, "src.zip"),
        ContentHash: aws.GetPayloadHash(zipBall),
    }
	return info,nil
	
}


func DeployEC2(AccessKey string, SecretKey string, bucket S3DeploymentInfo, service models.Service) ([]EC2InstanceInfo,error){
	client, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://ec2.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "ec2",
	})
	if err != nil {
		fmt.Println(err)
		return []EC2InstanceInfo{},err
	}
	
	buildScript := generateBuildScript(bucket, service,AccessKey, SecretKey)
	fmt.Println("Generated build script: ", buildScript)
	fmt.Println("Created EC2 client: ", client)
	params := map[string]string{
		"ImageId": "ami-01816d07b1128cd2d",
		"InstanceType": "t2.micro",
		"UserData": buildScript,
		"MinCount": "1",
		"MaxCount": "1",
	}
	fmt.Println("Creating instance")
	resp, err := client.RunInstances(context.Background(), params)
	fmt.Println(resp)
	var instances []EC2InstanceInfo
    for _, instance := range resp.Instances {
        newInstance := EC2InstanceInfo{
            InstanceID:     instance.InstanceId,
            PrivateIP:      instance.PrivateIpAddress,
            State:          instance.State.Name,
            Region:        "us-east-1",
        }
        instances = append(instances, newInstance)
    }
	return instances,nil;
}
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

func generateBuildScript(s3Info S3DeploymentInfo, service models.Service, accessKey, secretKey string, files map[string][]byte) string {
    // Detect buildpack from file map
    buildpack, err := DetectBuildPack(files)
    if err != nil {
        log.Printf("Warning: No specific buildpack detected, using default behavior")
        buildpack = &BuildPack{
            Name: "Generic",
            Packages: []string{},
            PreBuild: []string{},
            Build: service.Build,
            EnvVars: map[string]string{},
        }
    }

    // Your existing heading
    heading := `#!/bin/bash
    set -e...` // (keeping your existing heading)

    // Modify script to use buildpack info
    script := `
    log "Starting build process for %s application..."

    # Update system
    log "Updating system packages..."
    dnf update -y

    # Install AWS CLI and your existing setup...
    # (keeping your AWS setup section)

    # Install detected runtime packages
    log "Installing detected runtime packages..."
    dnf install -y %s

    # Set up environment variables
    log "Configuring environment..."
    %s

    # Run pre-build commands
    log "Running pre-build setup..."
    %s

    # Run build commands
    log "Running build process..."
    %s

    # Create systemd service
    log "Creating systemd service..."
    cat > /etc/systemd/system/%s.service << EOF
    [Unit]
    Description=%s Service (%s)
    After=network.target

    [Service]
    Type=simple
    User=root
    WorkingDirectory=/app/$EXTRACT_DIR/%s
    ExecStart=%s
    Restart=always
    Environment=PORT=%d
    %s

    [Install]
    WantedBy=multi-user.target
    EOF
    `

    // Create environment variables string
    var envVars strings.Builder
    for k, v := range buildpack.EnvVars {
        envVars.WriteString(fmt.Sprintf("export %s=%s\n", k, v))
    }

    formattedScript := heading + fmt.Sprintf(script,
        buildpack.Name,
        strings.Join(buildpack.Packages, " "),
        envVars.String(),
        strings.Join(buildpack.PreBuild, "\n"),
        strings.Join(buildpack.Build, " && "),
        service.Name,
        service.Name,
        buildpack.Name,
        service.Path,
        buildpack.Start,
        service.Port,
        // Convert env vars to systemd format
        strings.Join(func() []string {
            var envs []string
            for k, v := range buildpack.EnvVars {
                envs = append(envs, fmt.Sprintf("Environment=%s=%s", k, v))
            }
            return envs
        }(), "\n"),
    )

    return base64.StdEncoding.EncodeToString([]byte(formattedScript))
}
func ProvisionNetwork(AccessKey string, SecretKey string, projectName string ) ( string, string, error) {
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://elasticloadbalancing.amazonaws.com/",
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return "","",  err
	}
	ec2Client, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://ec2.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "ec2",})
		if err != nil {
			fmt.Println(err)
			return "","", err
		}
	 subnet1, subnet2,sgId, vpcId, err := ec2Client.GetAlbNetworkInfo(context.Background() )
	if err != nil {
		fmt.Println(err)
		return "","", err
	}
	/*targetArn, err := albClient.CreateTargetGroup(context.Background(), base+"-"+projectName+"-Byteport", vpcId)
	if err != nil {
		fmt.Println(err)
		return  "","",err
	}*/
	fmt.Println("VPC: ", vpcId);
	albInstance, err := albClient.CreateInternetApplicationLoadbalancer(context.Background(), projectName, sgId, subnet1, subnet2)
	if err != nil {
		fmt.Println(err)
		return "","", err
	}
	 loadBalancerArn := albInstance.CreateLoadBalancerResult.LoadBalancers.Member.LoadBalancerArn
	  
	 
	// for each service create targetgroup service-TG -> ALB Listener Rule Path /service/* -> service-TG
	
	fmt.Println("Hosted zone created successfully.")
	
	 
	return loadBalancerArn, vpcId,nil
	
}
func CreateALBListener(AccessKey string, SecretKey string, projectName string, loadBalancerArn string, vpcId string, instanceId string, port int  ) (string, error ){
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://elasticloadbalancing.amazonaws.com/",
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	targetArn, err := RegisterService(AccessKey, SecretKey, loadBalancerArn, projectName, "main", vpcId, instanceId, port)
	listenerResponse,err := albClient.CreateListener(context.Background(), projectName, loadBalancerArn, targetArn)
	 if err != nil {
		fmt.Println(err)
		return "",err
	}
	listenerArn := listenerResponse.CreateListenerResult.Listeners.Member.ListenerArn 
	return listenerArn,nil

}
func SetListenerRules(AccessKey string, SecretKey string, ListenerArn string, TargetArn string, serviceName string, priority int) (error){
	c, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://elasticloadbalancing.amazonaws.com/",
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return err}
	err = c.CreateListenerRule(context.Background(), ListenerArn, TargetArn, serviceName, priority)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func RegisterService(AccessKey string, SecretKey string, loadBalancerArn string, projectName string, serviceName string, vpcId string, instanceId string, port int) (string,error){
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://elasticloadbalancing.amazonaws.com/",
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return  "",err
	}
	targetArn, err := albClient.CreateTargetGroup(context.Background(), serviceName+"-"+projectName+"-Byteport", vpcId)
	if err != nil {
		fmt.Println(err)
		return  "",err
	}
	err =albClient.RegisterTarget(context.Background(), targetArn, instanceId, port )
	if err != nil {
		fmt.Println(err)
		return  "",err
	}
	

	return targetArn, nil;
}
func AddNewRecord(AccessKey string, SecretKey string, domainName string, zoneID string, projectName string, value string) (string, error) {
	c, err := r53.NewRoute53(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://route53.amazonaws.com/2013-04-01/hostedzone",
		Region: "us-east-1",
		Service: "route53",
	})
	if err != nil {
		fmt.Println(err)
		return "",err
	}

	err = c.CreateRecordSet(context.Background(), zoneID, domainName, "A", value, 300, projectName)
	if err != nil {
		fmt.Println(err)
		return "",err
	}
	fmt.Println("Record set created successfully.")
	return "Success",nil
	
}
func AwaitInitialization(AccessKey string, SecretKey string, instanceIDs []string) (error){
	fmt.Println("Waiting for instances to initialize...")
	c, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://ec2.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "ec2"})
	if err != nil {
		fmt.Println(err)
		return err }
		fmt.Println("Created EC2 client")
	err = c.WaitForEC2Running(instanceIDs, context.Background())
	if err != nil {
		fmt.Println(err)
		return err}
	fmt.Println("Instances initialized")
	return nil
}
 