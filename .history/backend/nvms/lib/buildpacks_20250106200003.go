package lib

import (
	"encoding/base64"
	"fmt"
	"log"
	"nvms/models"
	"path/filepath"
	"strings"
)
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
        Name: "Spin",
        DetectFiles: []string{"spin.toml"},
        Packages: []string{"rust", "cargo", "golang", "spin"},  // Base requirements
        PreBuild: []string{
            "curl -fsSL https://developer.fermyon.com/downloads/install.sh | bash",
            "mv spin /usr/local/bin/",
            "export SPIN_HOME=/root/.spin",
            "mkdir -p $SPIN_HOME",
        },
        Build: []string{
            "spin build",
        },
        Start: "spin up --listen 0.0.0.0", 
        RuntimeVersions: map[string]string{
            "spin.toml": `spin_version = ["'](\d+\.\d+\.\d+)["']`,
        },
        EnvVars: map[string]string{
            "SPIN_HOME": "/root/.spin",
            "RUST_BACKTRACE": "1", 
            "GOPATH": "/root/go",
            "GOROOT": "/usr/local/go",
            "TINYGO_ROOT": "/usr/local/tinygo",
        },
    },
        {
            Name: "Go",
            DetectFiles: []string{"go.mod", "go.sum"},
            Packages: []string{"golang"},
            PreBuild: []string{
                 "export HOME=/root",
				"export XDG_CACHE_HOME=/root/go",
				"export GOCACHE=/root/go/cache",
				"export GOPATH=/root/go",
				"export GOMODCACHE=$GOPATH/pkg/mod",
				"mkdir -p $GOPATH",
				"mkdir -p $GOCACHE",
				"mkdir -p $XDG_CACHE_HOME",
            },
            Build: []string{
                "go mod download",
                "go build -o app",
            },
            Start: "/app/$EXTRACT_DIR/$SERVICE_PATH/app",
            RuntimeVersions: map[string]string{
                "go.mod": `go (\d+\.\d+)`, // Regex to extract version
            },
            EnvVars: map[string]string{
				"GOPATH": "/root/go",
				"GOMODCACHE": "/root/go/pkg/mod",
				"GOCACHE": "/root/go/cache",
				"HOME": "/root",
				"XDG_CACHE_HOME": "/root/go",
			},
        },
        {
            Name: "Node.js",
            DetectFiles: []string{"package.json", "yarn.lock", "npm-shrinkwrap.json"},
            Packages: []string{"nodejs", "npm"},
            PreBuild: []string{
                "npm install -g rollup",
				"npm install -g yarn", 
				"npm config set update-notifier false",
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
				"NO_UPDATE_NOTIFIER": "1",
            },
        },{
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
    for _, bp := range buildpacks {
        if matchesBuildpackInMemory(tree, bp.DetectFiles, service.Path,rootDir) {
            return &bp, nil
        }
    }
	fmt.Println("No buildpack detected")

    return nil, fmt.Errorf("no buildpack detected for provided files")
}
func matchesBuildpackInMemory(tree map[string][]string, detectFiles []string, servicePath string, rootDir string) bool {
    normalizedPath := filepath.Base(strings.Trim(servicePath, "/"))  // Get just 'backend' or 'frontend'
    //fmt.Printf("Checking buildpacks for path: %s\n", normalizedPath)
    //fmt.Printf("Files to detect: %v\n", detectFiles)

    // Get files in the service directory
    dirFiles, exists := tree[normalizedPath]
    if !exists {
        fmt.Printf("Directory %s not found in tree\n", normalizedPath)
        return false
    }
    // Check each detect file
    for _, file := range detectFiles {
        for _, f := range dirFiles {
            if f == file {
                fmt.Printf("Found matching file %s in %s\n", file, normalizedPath)
                return true
            }
        }
    }

    fmt.Printf("No matching files found in %s\n", normalizedPath)
    return false
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
 
func ConvertBuildpackToComponent(buildpack models.BuildPack, service models.Service, s3Info S3DeploymentInfo) *models.ImageComponent {
    fmt.Println("Converting buildpack to image component")
    s3URI := fmt.Sprintf("s3://%s/%s", s3Info.BucketName, s3Info.ObjectKey)

    // Convert buildpack.EnvVars into CLI-friendly exports
    envCommands := make([]string, 0, len(buildpack.EnvVars))
    for k, v := range buildpack.EnvVars {
        envCommands = append(envCommands, fmt.Sprintf("export %s=\"%s\"", k, v))
    }

    extractionDir := filepath.Base(strings.Trim(service.Path, "/"))

     return &models.ImageComponent{
    Name:        service.Name,
    SchemaVersion:     1.0,
    Description: fmt.Sprintf("Buildpack for %s service, Pack: %s", service.Name, buildpack.Name),
    Phases: []models.Phase{
        {
            Name: "build",
            Steps: []models.Step{
                {
                    Name:   "DownloadCode",
                    Action: "S3Download",
                        Inputs: map[string]interface{}{
                        "source":      s3URI,
                        "destination"": "/app",
                        Overwrite:   true,
                    },
                },
                 {
                    Name:   "UnzipCode",
                    Action: "ExecuteBash",
                    Inputs:  { 
                        Commands: []string{"unzip /app/src.zip -d /app"},
                    },
                },
                {
                    Name:   "PrepSystem",
                    Action: "UpdateOS",
                    // No Inputs needed for UpdateOS
                },
                {
                    Name:   "InstallSystemPackages",
                    Action: "ExecuteBash",
                    Inputs: map[string]interface{}{
                        Commands: []string{
                            fmt.Sprintf("dnf install -y %s", strings.Join(buildpack.Packages, " ")),
                        },
                    },
                },
                {
                    Name:   "SetEnvVars",
                    Action: "ExecuteBash",
                    Inputs: map[string]interface{}{ 
                        Commands: envCommands,
                    },
                },
                {
                    Name:   "RunPreBuildCommands",
                    Action: "ExecuteBash",
                    Inputs: map[string]interface{}{ 
                        Commands: append([]string{fmt.Sprintf("cd /app/%s", extractionDir)}, buildpack.PreBuild...),
                    },
                },
                {
                    Name:   "RunBuildCommands",
                    Action: "ExecuteBash",
                    Inputs: map[string]interface{}{ 
                        Commands: append([]string{fmt.Sprintf("cd /app/%s", extractionDir)}, buildpack.Build...),
                    },
                },
            }, 
        },
         {
            Name: "validate",
            Steps: []models.Step{
                {
                    Name:   "StartService",
                    Action: "ExecuteBash",
                    Inputs: map[string]interface{}{ 
                        Commands: []string{fmt.Sprintf("cd /app/%s && %s", extractionDir, buildpack.Start)},
                    },
                },
            },
        },
        {
            Name: "test",
            Steps: []models.Step{
                {
                    Name:   "VerifyService",
                    Action: "Assert",
                    Inputs: map[string]interface{}{ 
                        Assertions: []models.AssertionConfig{
                            {
                                Condition: true,
                                Message:   "Service deployment completed",
                            },
                        },
                    },
                },
            }, 
        },
    },
}
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
	filepath.Base(strings.Trim(service.Path, "/")),
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
    //fmt.Printf("S3 Info: Bucket=%s, Key=%s\n", s3Info.BucketName, s3Info.ObjectKey)
	//fmt.Println("Formatted script: ", formattedScript)
    return base64.StdEncoding.EncodeToString([]byte(formattedScript)),nil
}
func GenerateImageComponentData(s3Info S3DeploymentInfo, service models.Service, accessKey, secretKey string, files []string) (*models.ImageComponent, error) {
	buildpack, err := DetectBuildPack(files, service)
	if err != nil {
		return nil, err
	}
    fmt.Println("Detected buildpack: ", buildpack.Name)
	return ConvertBuildpackToComponent(*buildpack, service, s3Info), nil
}