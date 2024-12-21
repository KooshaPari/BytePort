package lib

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
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


func DeployEC2(AccessKey string, SecretKey string, bucket S3DeploymentInfo, service models.Service, fileMap []string) ([]EC2InstanceInfo,error){
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
	
	buildScript,err := generateBuildScript(bucket, service,AccessKey, SecretKey, fileMap)
	if err != nil {
		fmt.Println("Error generating build script: ", err)
		return []EC2InstanceInfo{}, err
	}
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
func DetectBuildPack(files []string, service models.Service) (*models.BuildPack, error) {
    if(service.BuildPack != nil){ 
		fmt.Println("Service has buildpack")
        // check if buildpack has required fields
        if service.BuildPack.Name == "" || len(service.BuildPack.Build) == 0 || service.BuildPack.Start == "" || len(service.BuildPack.Packages) == 0 || len(service.BuildPack.PreBuild) == 0 {
			fmt.Println("Buildpack is missing required parameters")
            return nil, fmt.Errorf("buildpack is missing required parameters")
        }
        return service.BuildPack, nil
    }
    buildpacks := []models.BuildPack{
        {
            Name: "Go",
            DetectFiles: []string{"go.mod", "go.sum"},
            Packages: []string{"golang"},
            PreBuild: []string{
                "export GOPATH=/root/go",
                "export GOMODCACHE=$GOPATH/pkg/mod",
                "mkdir -p $GOPATH",
            },
            Build: []string{
                "go mod download",
                "go build -o app",
            },
            Start: "./app",
            RuntimeVersions: map[string]string{
                "go.mod": `go (\d+\.\d+)`, // Regex to extract version
            },
            EnvVars: map[string]string{
                "GOPATH": "/root/go",
                "GOMODCACHE": "/root/go/pkg/mod",
            },
        },
        {
            Name: "Node.js",
            DetectFiles: []string{"package.json", "yarn.lock", "npm-shrinkwrap.json"},
            Packages: []string{"nodejs", "npm"},
            PreBuild: []string{
                "npm install -g yarn", // If yarn.lock exists
            },
            Build: []string{
                "npm install",
                "npm run build",
            },
            Start: "npm start",
            RuntimeVersions: map[string]string{
                "package.json": `"node": "(\d+\.\d+\.\d+)"`,
                ".nvmrc": `^v?(\d+\.\d+\.\d+)$`,
            },
            EnvVars: map[string]string{
                "NPM_CONFIG_PRODUCTION": "true",
            },
        },
        {
            Name: "Python",
            DetectFiles: []string{"requirements.txt", "Pipfile", "pyproject.toml"},
            Packages: []string{"python3", "python3-pip", "python3-venv"},
            PreBuild: []string{
                "python3 -m venv venv",
                "source venv/bin/activate",
            },
            Build: []string{
                "pip install -r requirements.txt",
            },
            Start: "python app.py",
            RuntimeVersions: map[string]string{
                "runtime.txt": `python-(\d+\.\d+\.\d+)`,
                "Pipfile": `python_version = "(\d+\.\d+)"`,
            },
            EnvVars: map[string]string{
                "PYTHONPATH": "/app",
            },
        },
        {
            Name: "Java",
            DetectFiles: []string{"pom.xml", "build.gradle", ".mvn"},
            Packages: []string{"java-11-openjdk", "maven"},
            PreBuild: []string{},
            Build: []string{
                "mvn clean install",
            },
            Start: "java -jar target/*.jar",
            RuntimeVersions: map[string]string{
                "system.properties": `java.runtime.version=(\d+)`,
            },
            EnvVars: map[string]string{
                "JAVA_OPTS": "-Xmx300m -Xss512k -XX:CICompilerCount=2",
            },
        },
        {
            Name: "Ruby",
            DetectFiles: []string{"Gemfile", "config.ru", "Rakefile"},
            Packages: []string{"ruby", "ruby-devel", "gcc", "make"},
            PreBuild: []string{
                "gem install bundler",
            },
            Build: []string{
                "bundle install",
            },
            Start: "bundle exec ruby app.rb",
            RuntimeVersions: map[string]string{
                "Gemfile": `ruby ['\"](\d+\.\d+\.\d+)['\"]`,
                ".ruby-version": `^(\d+\.\d+\.\d+)`,
            },
            EnvVars: map[string]string{
                "RACK_ENV": "production",
            },
        },
        {
            Name: "PHP",
            DetectFiles: []string{"composer.json", "index.php", "artisan"},
            Packages: []string{"php", "php-fpm", "php-mysql", "composer"},
            PreBuild: []string{},
            Build: []string{
                "composer install --no-dev",
            },
            Start: "php-fpm",
            RuntimeVersions: map[string]string{
                "composer.json": `"php": ["']>=?(\d+\.\d+)`,
            },
            EnvVars: map[string]string{
                "PHP_FPM_PM": "dynamic",
            },
        },
        {
            Name: "Rust",
            DetectFiles: []string{"Cargo.toml", "Cargo.lock"},
            Packages: []string{"rust", "cargo"},
            PreBuild: []string{},
            Build: []string{
                "cargo build --release",
            },
            Start: "./target/release/app",
            RuntimeVersions: map[string]string{
                "rust-toolchain.toml": `channel = ["'](\d+\.\d+)["']`,
            },
            EnvVars: map[string]string{
                "RUST_BACKTRACE": "1",
            },
        },
    
        
    }
	fmt.Println("Checking buildpacks")
	tree, rootDir, err := AnalyzeBuildpackPaths(files)
	if err != nil {
		fmt.Println("Error analyzing buildpack paths: ", err)
		return nil, err
	}
	

    // Check files in memory instead of on disk
    for _, bp := range buildpacks {
        if matchesBuildpackInMemory(tree, bp.DetectFiles, service.Path,rootDir) {
            return &bp, nil
        }
    }
	fmt.Println("No buildpack detected")

    return nil, fmt.Errorf("no buildpack detected for provided files")
}
func matchesBuildpackInMemory(tree map[string][]string, detectFiles []string, servicePath string, rootDir string) bool {
    normalizedPath := strings.Trim(servicePath, "/")
    fmt.Println("Checking buildpacks for path: ", normalizedPath)
    fmt.Println("Root directory: ", rootDir)

    // Construct search path
    searchPath := filepath.Join(rootDir, normalizedPath)
    fmt.Println("Searching in path: ", searchPath)

    // Check for files in the directory
    for _, file := range detectFiles {
        dirFiles, exists := tree[searchPath]
        if exists {
            for _, f := range dirFiles {
                if f == file {
                    fmt.Println("Buildpack detected for file: ", file)
                    return true
                }
				else {
					
				}
            }
        }
    }

    fmt.Println("No buildpack detected for path: ", searchPath)
    return false
}
func generateBuildScript(s3Info S3DeploymentInfo, service models.Service, accessKey, secretKey string, files []string) (string, error ){
	fmt.Println("Getting Buildpack")
    buildpack, err := DetectBuildPack(files, service)
    if err != nil {
		fmt.Println("Error detecting buildpack: ", err)
        log.Printf("Warning: No specific buildpack detected, using default behavior")
        buildpack = &models.BuildPack{
            Name: "Generic",
            Packages: []string{},
            PreBuild: []string{},
            Build: service.Build,
            EnvVars:  map[string]string{},
			Start: strings.Join(service.Build, " && "),
			DetectFiles: []string{},
			RuntimeVersions: map[string]string{},
        }
		return "", err
    }
	fmt.Println("Got Buildpack: ", buildpack)
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
cat > /root/.aws/credentials << EOF
[default]
aws_access_key_id = %s
aws_secret_access_key = %s
region = us-east-1
EOF

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
	fmt.Println("Building script...")
	envVarsList := make([]string, 0, len(buildpack.EnvVars))
    for k, v := range buildpack.EnvVars {
        envVarsList = append(envVarsList, fmt.Sprintf("export %s=%s", k, v))
    }
    environmentVars := strings.Join(envVarsList, "\n")
    // Format script with actual values
    formattedScript := heading + fmt.Sprintf(script,
    buildpack.Name,            // %s for application type
    accessKey,                 // %s for AWS access key
    secretKey,                 // %s for AWS secret key
    s3Info.BucketName,        // %s for bucket name
    s3Info.ObjectKey,         // %s for object key
    service.Path,             // %s for service path (logging)
    service.Path,             // %s for service path (cd)
    strings.Join(buildpack.Packages, " "), // %s for packages
    environmentVars,         // %s for env vars
    strings.Join(buildpack.PreBuild, "\n"), // %s for prebuild
    strings.Join(buildpack.Build, " && "),  // %s for build commands
    service.Name,             // %s for service name
    service.Name,             // %s for service name in Description
    buildpack.Name,           // %s for buildpack name
    service.Path,             // %s for WorkingDirectory
    buildpack.Start,          // %s for ExecStart
    service.Port,             // %d for PORT
    strings.Join(func() []string {  // %s for systemd env vars
        var envs []string
        for k, v := range buildpack.EnvVars {
            envs = append(envs, fmt.Sprintf("Environment=%s=%s", k, v))
        }
        return envs
    }(), "\n"),
    service.Name,             // %s for enable
    service.Name,             // %s for start
)
    // Debug print the parameters (remove sensitive info)
    fmt.Printf("Service: %+v\n", service)
    fmt.Printf("Build Pack: %s\n",  buildpack )
    fmt.Printf("S3 Info: Bucket=%s, Key=%s\n", s3Info.BucketName, s3Info.ObjectKey)
	fmt.Println("Formatted script: ", formattedScript)
    return base64.StdEncoding.EncodeToString([]byte(formattedScript)),nil
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
 func AnalyzeBuildpackPaths(paths []string) (map[string ][]string,string, error) {
    // Extract root directory
    rootDir := findCommonPrefix(paths)
    if rootDir == "" {
        return nil, "", fmt.Errorf("no common root directory found")
    }

    // Create file tree
    tree := make(map[string][]string)
    for _, path := range paths {
        relativePath := strings.TrimPrefix(path, rootDir)
        dir := filepath.Dir(relativePath)
        tree[dir] = append(tree[dir], filepath.Base(relativePath))
    }

       return tree, rootDir,nil
    }

   
func findCommonPrefix(paths []string) string {
    if len(paths) == 0 {
        return ""
    }

    prefix := paths[0]
    for _, path := range paths[1:] {
        for !strings.HasPrefix(path, prefix) {
            prefix = prefix[:strings.LastIndex(prefix, "/")]
        }
    }
    return prefix
}

 

 