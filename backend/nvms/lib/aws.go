package lib

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"

	aws "nvms/lib/awspin"
	ec2 "nvms/lib/awspin/ec2"
	awsnet "nvms/lib/awspin/network"
	r53 "nvms/lib/awspin/network/route53"
	"nvms/lib/awspin/s3"
	"nvms/models"
	"strings"

	"github.com/google/uuid"
)

// getAWSRegion returns the AWS region from environment variable or defaults to us-east-1
func getAWSRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	return "us-east-1"
}

// ShellQuote returns a POSIX-shell-safe single-quoted version of s. Values
// supplied by users (env vars, service paths, names) flow into the EC2
// user-data script, so any unescaped single-quote or backslash in user
// input would be a shell-injection vector. Exported so it can be unit-tested
// from aws_test.go.
//
// The algorithm is the standard "'foo' becomes '\”foo'\”" idiom: wrap the
// entire string in single quotes, but exit and re-enter single quotes around
// every embedded single quote (replacing each ' with `'\”`).
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

var AWSEndpointBase string = "https://%s." + getAWSRegion() + ".amazonaws.com" /* "http://localhost.localstack.cloud:4566"*/
func PushToS3(zipBall []byte, AccessKey string, SecretKey string, ProjectName string) (S3DeploymentInfo, error) {
	log.Println("Uploading to S3...")
	cfg := aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("s3"),
		Region:          getAWSRegion(),
		Service:         "s3",
	}
	ctx := context.Background()
	s3Client, err := s3.NewS3(cfg)
	if err != nil {
		log.Printf("S3 client creation failed: %v\n", err)
		return S3DeploymentInfo{}, err
	}
	log.Println("S3 client created")
	bucketName := strings.ToLower(ProjectName) + "-bytebucket-" + uuid.New().String()
	err = s3Client.CreateBucket(ctx, bucketName)
	if err != nil {
		log.Printf("Bucket creation failed: %v\n", err)
		return S3DeploymentInfo{}, err
	}

	log.Printf("Bucket created: %s\n", bucketName)
	err = s3Client.PutObject(ctx, bucketName, "src.zip", zipBall)
	if err != nil {
		log.Printf("PutObject failed: %v\n", err)
		return S3DeploymentInfo{}, err
	}
	log.Printf("Uploaded to S3: %s\n", bucketName)
	// return uri/bucket name for later use

	info := S3DeploymentInfo{
		BucketName: bucketName,
		ObjectKey:  "src.zip",
		Region:     getAWSRegion(),
		BucketARN:  fmt.Sprintf("arn:aws:s3:::%s", bucketName),
		//ObjectURL:   fmt.Sprintf("http://localhost:4566/%s/%s", bucketName, "src.zip"),
		ObjectURL:   fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, "src.zip"),
		ContentHash: aws.GetPayloadHash(zipBall),
	}
	return info, nil

}

// DeployEC2 launches an EC2 instance running the build pipeline for `service`.
//
// Security: when `iamInstanceProfile` is non-empty the EC2 instance is launched
// with an IAM instance profile attached, and the user-data script MUST NOT
// embed any static AWS credentials. The build script will then read temporary
// credentials from the EC2 instance metadata service (IMDSv2). Callers SHOULD
// always pass an IAM instance profile ARN; the legacy `AccessKey`/`SecretKey`
// path exists only for LocalStack development.
func DeployEC2(AccessKey string, SecretKey string, bucket S3DeploymentInfo, service models.Service, fileMap []string, iamInstanceProfile string) ([]EC2InstanceInfo, error) {
	client, err := ec2.NewEC2(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("ec2"),
		Region:          getAWSRegion(),
		Service:         "ec2",
	})
	if err != nil {
		log.Printf("EC2 client creation failed: %v\n", err)
		return []EC2InstanceInfo{}, err
	}

	buildScript, err := generateBuildScript(bucket, service, AccessKey, SecretKey, fileMap, iamInstanceProfile)
	if err != nil {
		log.Printf("Error generating build script: %v\n", err)
		return []EC2InstanceInfo{}, err
	}
	log.Println("EC2 client created")
	params := map[string]string{
		"ImageId":      "ami-01816d07b1128cd2d",
		"InstanceType": "t2.micro",
		"UserData":     buildScript,
		"MinCount":     "1",
		"MaxCount":     "1",
	}
	// If we have an IAM instance profile, attach it. The build script will
	// detect this case and NOT embed static credentials in user-data.
	if iamInstanceProfile != "" {
		params["IamInstanceProfile.Arn"] = iamInstanceProfile
		log.Println("EC2 launch will use IAM instance profile (no static creds in user-data)")
	}
	log.Println("Creating EC2 instance")
	resp, err := client.RunInstances(context.Background(), params)
	//fmt.Println(resp)
	var instances []EC2InstanceInfo
	for _, instance := range resp.Instances {
		newInstance := EC2InstanceInfo{
			InstanceID: instance.InstanceId,
			PrivateIP:  instance.PrivateIpAddress,
			State:      instance.State.Name,
			Region:     getAWSRegion(),
		}
		instances = append(instances, newInstance)
	}
	return instances, nil
}

// generateBuildScript returns a base64-encoded shell script suitable for use
// as EC2 user-data.
//
// Security: when `iamInstanceProfile` is non-empty the script does NOT embed
// the `accessKey`/`secretKey` parameters; AWS credentials are obtained from
// the EC2 instance metadata service (IMDSv2) via the attached IAM instance
// profile. Static credentials are still supported for LocalStack development
// but should never be used in production deployments.
func generateBuildScript(s3Info S3DeploymentInfo, service models.Service, accessKey, secretKey string, files []string, iamInstanceProfile string) (string, error) {
	log.Println("Getting Buildpack")
	buildpack, err := DetectBuildPack(files, service)
	if err != nil {
		log.Printf("Error detecting buildpack: %v\n", err)
		log.Printf("Warning: No specific buildpack detected, using default behavior")
		buildpack = &models.BuildPack{
			Name:            "Generic",
			Packages:        []string{},
			PreBuild:        []string{},
			Build:           service.Build,
			EnvVars:         map[string]string{},
			Start:           strings.Join(service.Build, " && "),
			DetectFiles:     []string{},
			RuntimeVersions: map[string]string{},
		}
		return "", err
	}
	//fmt.Println("Got Buildpack: ", buildpack)
	heading := `#!/bin/bash
set -e

# Configure logging
exec 1> >(logger -s -t $(basename $0)) 2>&1
BUILD_LOG="/var/log/user-data-build.log"
touch $BUILD_LOG
chmod 644 $BUILD_LOG

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a $BUILD_LOG
}`
	script := `

log "Starting build process for %s application..."

# Update system
log "Updating system packages..."
dnf update -y

# Install AWS CLI and required tools
log "Installing required tools..."
dnf install -y unzip tar gzip

# Install AWS CLI v2
log "Installing AWS CLI..."
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install
rm -f awscliv2.zip
rm -rf aws/

# Configure AWS credentials
log "Configuring AWS credentials..."
mkdir -p /root/.aws
if [ -n "${USE_IAM_INSTANCE_PROFILE:-}" ]; then
    # Production: rely on the EC2 IAM instance profile. The CLI will fetch
    # temporary credentials from IMDSv2. Do NOT write static credentials.
    log "Using IAM instance profile credentials (IMDSv2)"
    cat > /root/.aws/config << EOF
[default]
region = us-east-1
EOF
else
    # LocalStack / dev fallback only. Never use in production.
    log "WARNING: using static AWS credentials (LocalStack/dev only)"
    cat > /root/.aws/credentials << EOF
[default]
aws_access_key_id = %s
aws_secret_access_key = %s
region = us-east-1
EOF
fi

# Verify AWS configuration
aws configure list

# Create working directory
log "Creating working directory..."
mkdir -p /app
cd /app

# Download code from S3
log "Downloading code from S3..."
aws s3 cp s3://%s/%s src.zip

# Unzip the code
log "Extracting code..."
unzip src.zip
rm src.zip

# Find the actual directory
SERVICE_PATH=%s
EXTRACT_DIR=$(ls -d */ | head -n 1)
cd "$EXTRACT_DIR"

# Navigate to service directory
log "Navigating to service directory: %s"
cd %s
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


# Start service
log "Starting service..."
systemctl daemon-reload
systemctl enable %s
systemctl start %s

log "Build and deployment complete!"
`
	log.Println("Building script...")

	envVarsList := make([]string, 0, len(buildpack.EnvVars))
	for k, v := range buildpack.EnvVars {
		envVarsList = append(envVarsList, fmt.Sprintf("export %s=%s", ShellQuote(k), ShellQuote(v)))
	}
	environmentVars := strings.Join(envVarsList, "\n")

	// Pre-quote all user-supplied values that flow into the shell script.
	// service.Name / service.Path / s3Info.BucketName / s3Info.ObjectKey are
	// supplied by end users via the NVMS manifest, so they must always be
	// passed through ShellQuote.
	quotedServiceName := ShellQuote(service.Name)
	quotedServicePath := ShellQuote(service.Path)
	quotedBucket := ShellQuote(s3Info.BucketName)
	quotedObjectKey := ShellQuote(s3Info.ObjectKey)

	systemdEnvVars := strings.Join(func() []string {
		var envs []string
		for k, v := range buildpack.EnvVars {
			envs = append(envs, fmt.Sprintf("Environment=%s=%s", k, ShellQuote(v)))
		}
		return envs
	}(), "\n")

	buildScriptFmt := func(accessKeyPlaceholder, secretKeyPlaceholder string) string {
		return fmt.Sprintf(script,
			ShellQuote(buildpack.Name), // %s for application type (quoted)
			accessKeyPlaceholder,       // %s for AWS access key
			secretKeyPlaceholder,       // %s for AWS secret key
			quotedBucket,               // %s for bucket name (quoted)
			quotedObjectKey,            // %s for object key (quoted)
			ShellQuote(filepath.Base(strings.Trim(service.Path, "/"))), // %s for service path basename (quoted)
			quotedServicePath,                      // %s for service path (logging, quoted)
			quotedServicePath,                      // %s for service path (cd, quoted)
			strings.Join(buildpack.Packages, " "),  // %s for packages
			environmentVars,                        // %s for env vars (already shell-escaped)
			strings.Join(buildpack.PreBuild, "\n"), // %s for prebuild
			strings.Join(buildpack.Build, " && "),  // %s for build commands
			quotedServiceName,                      // %s for service name (quoted)
			quotedServiceName,                      // %s for service name in Description (quoted)
			ShellQuote(buildpack.Name),             // %s for buildpack name (quoted)
			quotedServicePath,                      // %s for WorkingDirectory (quoted)
			ShellQuote(buildpack.Start),            // %s for ExecStart (quoted)
			service.Port,                           // %d for PORT
			systemdEnvVars,                         // %s for systemd env vars (already shell-escaped)
			quotedServiceName,                      // %s for enable (quoted)
			quotedServiceName,                      // %s for start (quoted)
		)
	}

	// Build script body. When an IAM instance profile is in use, the
	// credential placeholders in the script template are NOT substituted, so
	// the corresponding %s slots are filled with empty strings. The conditional
	// block in the script then takes the IMDSv2 path.
	var formattedScript string
	if iamInstanceProfile != "" {
		// Production path: no static credentials. The %s slots that would have
		// received accessKey/secretKey are filled with "" (ignored at runtime).
		formattedScript = heading + "export USE_IAM_INSTANCE_PROFILE=1\n" + buildScriptFmt("", "")
	} else {
		// Legacy / LocalStack path: embed static credentials. Will print a
		// runtime warning to /var/log/user-data-build.log when used.
		formattedScript = heading + buildScriptFmt(ShellQuote(accessKey), ShellQuote(secretKey))
	}
	// Debug: log service and buildpack info
	log.Printf("Service: %+v\n", service)
	log.Printf("Build Pack: %s\n", buildpack)
	return base64.StdEncoding.EncodeToString([]byte(formattedScript)), nil
}

func ProvisionNetwork(AccessKey string, SecretKey string, projectName string) (*awsnet.CreateLoadBalancerResponse, string, string, error) {
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("elasticloadbalancing"),
		Region:          "us-east-1",
		Service:         "elasticloadbalancing",
	})
	if err != nil {
		log.Printf("ALB client creation failed: %v\n", err)
		return nil, "", "", err
	}
	ec2Client, err := ec2.NewEC2(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("ec2"),
		Region:          "us-east-1",
		Service:         "ec2"})
	if err != nil {
		log.Printf("EC2 client creation failed: %v\n", err)
		return nil, "", "", err
	}
	subnet1, subnet2, sgId, vpcId, err := ec2Client.GetAlbNetworkInfo(context.Background())
	if err != nil {
		log.Printf("GetAlbNetworkInfo failed: %v\n", err)
		return nil, "", "", err
	}
	/*targetArn, err := albClient.CreateTargetGroup(context.Background(), base+"-"+projectName+"-Byteport", vpcId)
	if err != nil {
		log.Printf("CreateTargetGroup failed: %v\n", err)
		return  "","",err
	}*/
	//fmt.Println("VPC: ", vpcId);
	albInstance, err := albClient.CreateInternetApplicationLoadbalancer(context.Background(), projectName, sgId, subnet1, subnet2)
	if err != nil {
		log.Printf("CreateInternetApplicationLoadbalancer failed: %v\n", err)
		return nil, "", "", err
	}
	//loadBalancerArn := albInstance.CreateLoadBalancerResult.LoadBalancers.Member.LoadBalancerArn
	publicDNS := albInstance.CreateLoadBalancerResult.LoadBalancers.Member.DNSName

	// for each service create targetgroup service-TG -> ALB Listener Rule Path /service/* -> service-TG

	log.Printf("ALB created with DNS: %s\n", publicDNS)

	return albInstance, vpcId, publicDNS, nil

}
func CreateALBListener(AccessKey string, SecretKey string, projectName string, loadBalancerArn string, vpcId string, instanceId string, port int) (string, string, error) {
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("elasticloadbalancing"),
		Region:          "us-east-1",
		Service:         "elasticloadbalancing",
	})
	targetArn, err := RegisterService(AccessKey, SecretKey, loadBalancerArn, projectName, "main", vpcId, instanceId, port)
	listenerResponse, err := albClient.CreateListener(context.Background(), projectName, loadBalancerArn, targetArn)
	if err != nil {
		log.Printf("CreateListener failed: %v\n", err)
		return "", "", err
	}
	listenerArn := listenerResponse.CreateListenerResult.Listeners.Member.ListenerArn
	log.Printf("ALB Listener created: %s\n", listenerArn)
	return listenerArn, targetArn, nil

}
func SetListenerRules(AccessKey string, SecretKey string, ListenerArn string, TargetArn string, serviceName string, priority int) error {
	c, err := awsnet.NewALB(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("elasticloadbalancing"),
		Region:          "us-east-1",
		Service:         "elasticloadbalancing",
	})
	if err != nil {
		log.Printf("ALB client creation failed: %v\n", err)
		return err
	}
	err = c.CreateListenerRule(context.Background(), ListenerArn, TargetArn, serviceName, priority)
	if err != nil {
		log.Printf("CreateListenerRule failed: %v\n", err)
		return err
	}
	return nil
}
func RegisterService(AccessKey string, SecretKey string, loadBalancerArn string, projectName string, serviceName string, vpcId string, instanceId string, port int) (string, error) {
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("elasticloadbalancing"),
		Region:          "us-east-1",
		Service:         "elasticloadbalancing",
	})
	if err != nil {
		log.Printf("ALB client creation failed: %v\n", err)
		return "", err
	}
	targetArn, err := albClient.CreateTargetGroup(context.Background(), serviceName+"-"+projectName+"-Byteport", vpcId)
	if err != nil {
		log.Printf("CreateTargetGroup failed: %v\n", err)
		return "", err
	}
	err = albClient.RegisterTarget(context.Background(), targetArn, instanceId, port)
	if err != nil {
		log.Printf("RegisterTarget failed: %v\n", err)
		return "", err
	}
	log.Printf("Service registered: %s\n", targetArn)
	return targetArn, nil
}
func AddNewRecord(AccessKey string, SecretKey string, domainName string, zoneID string, projectName string, value string) (string, error) {
	c, err := r53.NewRoute53(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("route53"),
		Region:          "us-east-1",
		Service:         "route53",
	})
	if err != nil {
		log.Printf("Route53 client creation failed: %v\n", err)
		return "", err
	}

	err = c.CreateRecordSet(context.Background(), zoneID, domainName, "A", value, 300, projectName)
	if err != nil {
		log.Printf("CreateRecordSet failed: %v\n", err)
		return "", err
	}
	log.Println("Record set created successfully")
	return "Success", nil

}
func AwaitInitialization(AccessKey string, SecretKey string, instanceIDs []string) error {
	log.Println("Waiting for instances to initialize...")
	c, err := ec2.NewEC2(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("ec2"),
		Region:          "us-east-1",
		Service:         "ec2",
	})
	if err != nil {
		log.Printf("EC2 client creation failed: %v\n", err)
		return err
	}
	log.Println("EC2 client created")
	err = c.WaitForEC2Running(instanceIDs, context.Background())
	if err != nil {
		log.Printf("WaitForEC2Running failed: %v\n", err)
		return err
	}
	log.Println("Instances initialized")
	return nil
}

func TerminateS3(resource models.AWSResource, AccessKey string, SecretKey string) error {
	c, err := s3.NewS3(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("s3"),
		Region:          "us-east-1",
		Service:         "s3",
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.DeleteBucket(context.Background(), resource.ID)
	if err != nil {
		err = c.DeleteObject(context.Background(), resource.ID, "src.zip")
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = c.DeleteBucket(context.Background(), resource.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return err
	}
	fmt.Println("Record set created successfully.")
	return nil
}

func TerminateEC2(resource models.AWSResource, AccessKey string, SecretKey string) error {
	c, err := ec2.NewEC2(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		Endpoint:        getServiceEndpoint("ec2"),
		Region:          "us-east-1",
		Service:         "ec2",
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.TerminateInstances(context.Background(), []string{resource.ID})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Record set created successfully.")

	return nil
}

func TerminateALB(resource models.AWSResource, AccessKey string, SecretKey string) error {
	c, err := awsnet.NewALB(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("elasticloadbalancing"),
		Region:          "us-east-1",
		Service:         "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.DeleteLoadbalancer(context.Background(), resource.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Record set created successfully.")
	return nil
}

func TerminateTargetGroup(resource models.AWSResource, AccessKey string, SecretKey string) error {
	c, err := awsnet.NewALB(aws.Config{
		AccessKeyId:     AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken:    "",
		Endpoint:        getServiceEndpoint("elasticloadbalancing"),
		Region:          "us-east-1",
		Service:         "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.DeleteTargetGroup(context.Background(), resource.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Record set created successfully.")
	return nil
}
