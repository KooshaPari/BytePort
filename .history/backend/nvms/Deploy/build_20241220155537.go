package deploy

import (
	"fmt"

	"nvms/models"
	"strings"
)


func GenerateBuildScript(buildpack *models.BuildPack, service models.Service) (string, error) {
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