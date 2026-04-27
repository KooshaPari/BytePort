package deploy

type BuildPack struct {
    Name         string
    DetectFiles  []string         // Files that indicate this buildpack should be used
    Packages     []string         // System packages needed
    PreBuild     []string         // Commands to run before building
    Build        []string         // Build commands
    Start        string           // Command to start the application
    RuntimeVersions map[string]string // Maps language version files to install commands
    EnvVars      map[string]string   // Required environment variables
}

func DetectBuildPack(servicePath string) (*BuildPack, error) {
    buildpacks := []BuildPack{
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

    for _, bp := range buildpacks {
        if matchesBuildpack(servicePath, bp.DetectFiles) {
            return &bp, nil
        }
    }

    return nil, fmt.Errorf("no buildpack detected for service at %s", servicePath)
}

func matchesBuildpack(path string, detectFiles []string) bool {
    for _, file := range detectFiles {
        if fileExists(filepath.Join(path, file)) {
            return true
        }
    }
    return false
}

func GenerateBuildScript(buildpack *BuildPack, service models.Service) (string, error) {
    // Base system setup
    script := `#!/bin/bash
set -e

# Configure logging
exec 1> >(logger -s -t $(basename $0)) 2>&1
BUILD_LOG="/var/log/user-data-build.log"
touch $BUILD_LOG
chmod 644 $BUILD_LOG

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a $BUILD_LOG
}
`
    // Install required packages
    script += fmt.Sprintf(`
log "Installing required packages for %s..."
dnf install -y %s
`, buildpack.Name, strings.Join(buildpack.Packages, " "))

    // Setup environment variables
    for k, v := range buildpack.EnvVars {
        script += fmt.Sprintf("export %s=%s\n", k, v)
    }

    // Pre-build setup
    script += "\nlog 'Running pre-build setup...'\n"
    for _, cmd := range buildpack.PreBuild {
        script += cmd + "\n"
    }

    // Build process
    script += "\nlog 'Running build process...'\n"
    for _, cmd := range buildpack.Build {
        script += cmd + "\n"
    }

    // Create systemd service
    script += fmt.Sprintf(`
log "Creating systemd service..."
cat > /etc/systemd/system/%s.service << EOF
[Unit]
Description=%s Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s
Restart=always
`, service.Name, service.Name, service.Path, buildpack.Start)

    // Add environment variables to systemd service
    for k, v := range buildpack.EnvVars {
        script += fmt.Sprintf("Environment=%s=%s\n", k, v)
    }

    script += `
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

    return script, nil
}