# BytePort Windows Setup Script
# Run as Administrator

param(
    [string]$Domain = "yourdomain.com",
    [string]$ProjectsPath = "C:\BytePort",
    [string]$TunnelName = "byteport-main",
    [ValidateSet("podman", "wsl-containers")]
    [string]$ContainerRuntime = "podman",
    [switch]$SkipCloudflare = $false
)

Write-Host "🚀 Setting up BytePort Windows Server..." -ForegroundColor Green

# Create directory structure
Write-Host "📁 Creating directory structure..." -ForegroundColor Yellow
$directories = @(
    "$ProjectsPath\projects",
    "$ProjectsPath\tunnels",
    "$ProjectsPath\nginx",
    "$ProjectsPath\logs",
    "$ProjectsPath\backups"
)

foreach ($dir in $directories) {
    if (!(Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force
        Write-Host "Created: $dir" -ForegroundColor Gray
    }
}

# Require Chocolatey if not present
if (!(Get-Command choco -ErrorAction SilentlyContinue)) {
    Write-Host "📦 Chocolatey is required before continuing." -ForegroundColor Yellow
    Write-Host "Install it from https://chocolatey.org/install, reopen PowerShell, then rerun this script." -ForegroundColor White
    exit 1
}

# Install required tools
Write-Host "🛠️ Installing required tools..." -ForegroundColor Yellow
$tools = @(
    "git",
    "golang",
    "nodejs"
)

if (!$SkipCloudflare) {
    # Install cloudflared manually since it's not in chocolatey
    Write-Host "Installing cloudflared..." -ForegroundColor Gray
    $cloudflaredUrl = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe"
    $cloudflaredPath = "$env:ProgramFiles\cloudflared\cloudflared.exe"

    if (!(Test-Path $cloudflaredPath)) {
        New-Item -ItemType Directory -Path "$env:ProgramFiles\cloudflared" -Force
        Invoke-WebRequest -Uri $cloudflaredUrl -OutFile $cloudflaredPath

        # Add to PATH
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
        if ($currentPath -notlike "*$env:ProgramFiles\cloudflared*") {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$env:ProgramFiles\cloudflared", "Machine")
        }
    }
}

foreach ($tool in $tools) {
    Write-Host "Installing $tool..." -ForegroundColor Gray
    choco install $tool -y
}

# Container lifecycle is delegated to the selected Podman/WSL Containers
# adapter.  This setup script deliberately does not install or invoke a
# competing container runtime; capability probing and lifecycle receipts are
# owned by NanoVMS/PhenoCompose.
$runtimeCommand = if ($ContainerRuntime -eq "podman") { "podman.exe" } else { "wslc.exe" }
if (!(Get-Command $runtimeCommand -ErrorAction SilentlyContinue)) {
    Write-Host "Warning: $runtimeCommand was not found. Install/configure it before starting services." -ForegroundColor Yellow
} else {
    Write-Host "Using container runtime adapter: $ContainerRuntime ($runtimeCommand)" -ForegroundColor Gray
}

# Configure environment variables
Write-Host "⚙️ Setting up environment variables..." -ForegroundColor Yellow

$envVars = @{
    "BYTEPORT_ROOT" = $ProjectsPath
    "BYTEPORT_DOMAIN" = $Domain
    "BYTEPORT_API_PORT" = "8081"
    "BYTEPORT_NVMS_PORT" = "3000"
    "BYTEPORT_FRONTEND_PORT" = "5173"
    "CONTAINER_RUNTIME" = $ContainerRuntime
    "CONTAINER_NETWORK" = "byteport-network"
    "TUNNEL_CONFIG_PATH" = "$ProjectsPath\tunnels"
    "TUNNEL_NAME" = $TunnelName
    "PROJECTS_PATH" = "$ProjectsPath\projects"
}

foreach ($var in $envVars.GetEnumerator()) {
    [Environment]::SetEnvironmentVariable($var.Key, $var.Value, "Machine")
    Write-Host "Set $($var.Key) = $($var.Value)" -ForegroundColor Gray
}

# Create .env file for local development
$envContent = @"
# BytePort Windows Configuration
BYTEPORT_ROOT=$ProjectsPath
BYTEPORT_DOMAIN=$Domain
BYTEPORT_API_PORT=8081
BYTEPORT_NVMS_PORT=3000
BYTEPORT_FRONTEND_PORT=5173
CONTAINER_RUNTIME=$ContainerRuntime
CONTAINER_NETWORK=byteport-network
TUNNEL_CONFIG_PATH=$ProjectsPath\tunnels
TUNNEL_NAME=$TunnelName
PROJECTS_PATH=$ProjectsPath\projects
"@

$envContent | Out-File -FilePath "$ProjectsPath\.env" -Encoding UTF8

# Configure Cloudflare Tunnel (if not skipped)
if (!$SkipCloudflare) {
    Write-Host "☁️ Configuring Cloudflare Tunnel..." -ForegroundColor Yellow

    Write-Host "Please run the following commands manually after this script completes:" -ForegroundColor Red
    Write-Host "1. cloudflared tunnel login" -ForegroundColor White
    Write-Host "2. cloudflared tunnel create $TunnelName" -ForegroundColor White
    Write-Host "3. Copy the tunnel credentials to $ProjectsPath\tunnels\" -ForegroundColor White

    # Create sample tunnel config
    $tunnelConfig = @"
tunnel: YOUR_TUNNEL_ID
credentials-file: $ProjectsPath\tunnels\YOUR_TUNNEL_ID.json

ingress:
  - hostname: $Domain
    service: http://localhost:8080
  - hostname: "*.$Domain"
    service: http://localhost:8080
  - service: http_status:404

logfile: $ProjectsPath\logs\tunnel.log
"@

    $tunnelConfig | Out-File -FilePath "$ProjectsPath\tunnels\config-template.yml" -Encoding UTF8
    Write-Host "Created tunnel config template at $ProjectsPath\tunnels\config-template.yml" -ForegroundColor Gray
}

# Create service management scripts.  The generated manager records the
# exact root PID, command marker, and start time for every service.  Stop only
# acts on a still-matching tracked process tree; it never searches by a
# process image name and never terminates unrelated agent/Codex processes.
Write-Host "🔧 Creating service management scripts..." -ForegroundColor Yellow

$escapedProjectsPath = $ProjectsPath.Replace("'", "''")
$serviceManagerTemplate = @'
[CmdletBinding()]
param(
    [ValidateSet('start', 'stop')]
    [string]$Action = 'start'
)

$ErrorActionPreference = 'Stop'
$Root = '__PROJECTS_PATH__'
$PidDirectory = Join-Path $Root '.byteport-pids'
$ProtectedProcessPattern = '(?i)(^|[\\/ _-])(codex(?:\.exe)?|codex-code-mode-host(?:\.exe)?|cursor(?:\.exe)?|claude(?:\.exe)?|claude-code|agent-runner|agent-host)([\\/ _.-]|$)'

$Services = @(
    [pscustomobject]@{
        Name = 'byteport-api'
        FilePath = 'go.exe'
        Arguments = @('run', 'main.go')
        WorkingDirectory = Join-Path $Root 'backend\byteport'
        CommandMarker = 'go run main.go'
    }
    [pscustomobject]@{
        Name = 'nvms'
        FilePath = 'go.exe'
        Arguments = @('run', 'main.go')
        WorkingDirectory = Join-Path $Root 'backend\nvms'
        CommandMarker = 'go run main.go'
    }
    [pscustomobject]@{
        Name = 'frontend'
        FilePath = 'npm.cmd'
        Arguments = @('run', 'dev')
        WorkingDirectory = Join-Path $Root 'frontend\web'
        CommandMarker = 'npm run dev'
    }
)

function Get-ProcessCommandLine {
    param([Parameter(Mandatory)][int]$Pid)

    $process = Get-CimInstance Win32_Process -Filter "ProcessId=$Pid" -ErrorAction SilentlyContinue
    if ($null -eq $process) { return $null }
    return [string]$process.CommandLine
}

function Get-DescendantProcessIds {
    param([Parameter(Mandatory)][int]$ParentPid)

    $children = @(Get-CimInstance Win32_Process -Filter "ParentProcessId=$ParentPid" -ErrorAction SilentlyContinue)
    foreach ($child in $children) {
        [int]$child.ProcessId
        Get-DescendantProcessIds -ParentPid ([int]$child.ProcessId)
    }
}

function Assert-TrackedTreeSafe {
    param([Parameter(Mandatory)][int[]]$ProcessIds)

    foreach ($candidatePid in $ProcessIds) {
        $commandLine = Get-ProcessCommandLine -Pid $candidatePid
        if ([string]::IsNullOrWhiteSpace($commandLine)) {
            throw "Refusing to stop PID ${candidatePid}: unable to verify command identity"
        }
        if ($commandLine -match $ProtectedProcessPattern) {
            throw "Refusing to stop protected agent/Codex PID $candidatePid"
        }
    }
}

function Get-RecordPath {
    param([Parameter(Mandatory)][string]$Name)

    return (Join-Path $PidDirectory "$Name.json")
}

function Read-TrackedProcess {
    param([Parameter(Mandatory)]$Service)

    $recordPath = Get-RecordPath -Name $Service.Name
    if (!(Test-Path -LiteralPath $recordPath)) { return $null }

    try {
        $record = Get-Content -LiteralPath $recordPath -Raw | ConvertFrom-Json
        $pid = [int]$record.Pid
    } catch {
        throw "Refusing to act on malformed PID record: $recordPath"
    }

    $commandLine = Get-ProcessCommandLine -Pid $pid
    if ([string]::IsNullOrWhiteSpace($commandLine)) {
        Remove-Item -LiteralPath $recordPath -Force
        return $null
    }

    $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
    if ($null -eq $process) {
        Remove-Item -LiteralPath $recordPath -Force
        return $null
    }

    $startedUtc = $process.StartTime.ToUniversalTime().ToString('o')
    if ($startedUtc -ne [string]$record.StartedUtc -or
        $commandLine -notlike "*$($Service.CommandMarker)*") {
        throw "Refusing to stop PID $pid for $($Service.Name): tracked identity no longer matches"
    }

    return [pscustomobject]@{
        Pid = $pid
        RecordPath = $recordPath
        CommandLine = $commandLine
    }
}

function Start-TrackedService {
    param([Parameter(Mandatory)]$Service)

    if (!(Test-Path -LiteralPath $Service.WorkingDirectory)) {
        throw "Service directory does not exist: $($Service.WorkingDirectory)"
    }

    $existing = Read-TrackedProcess -Service $Service
    if ($null -ne $existing) {
        throw "$($Service.Name) is already tracked at PID $($existing.Pid); stop it explicitly first"
    }

    $process = Start-Process -FilePath $Service.FilePath -ArgumentList $Service.Arguments -WorkingDirectory $Service.WorkingDirectory -PassThru
    $startedUtc = $process.StartTime.ToUniversalTime().ToString('o')
    [pscustomobject]@{
        Pid = $process.Id
        StartedUtc = $startedUtc
        CommandMarker = $Service.CommandMarker
    } | ConvertTo-Json | Set-Content -LiteralPath (Get-RecordPath -Name $Service.Name) -Encoding UTF8
    Write-Host "Started $($Service.Name) (PID $($process.Id))"
}

function Stop-TrackedService {
    param([Parameter(Mandatory)]$Service)

    $tracked = Read-TrackedProcess -Service $Service
    if ($null -eq $tracked) {
        Write-Host "$($Service.Name) is not running"
        return
    }

    # The identity check above is the safety boundary.  Descendants are
    # collected by parent PID, so no global image-name kill is possible.  The
    # second guard makes agent/Codex targeting fail closed even if a tracked
    # service tree was tampered with after it started.
    $descendants = @(Get-DescendantProcessIds -ParentPid $tracked.Pid)
    [array]::Reverse($descendants)
    Assert-TrackedTreeSafe -ProcessIds (@($tracked.Pid) + $descendants)
    foreach ($childPid in $descendants) {
        Stop-Process -Id $childPid -Force -ErrorAction SilentlyContinue
    }
    Stop-Process -Id $tracked.Pid -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $tracked.RecordPath -Force
    Write-Host "Stopped $($Service.Name) (tracked PID $($tracked.Pid))"
}

New-Item -ItemType Directory -Path $PidDirectory -Force | Out-Null
foreach ($service in $Services) {
    if ($Action -eq 'start') {
        Start-TrackedService -Service $service
    } else {
        Stop-TrackedService -Service $service
    }
}
'@
$serviceManager = $serviceManagerTemplate.Replace('__PROJECTS_PATH__', $escapedProjectsPath)
$serviceManager | Out-File -FilePath "$ProjectsPath\byteport-services.ps1" -Encoding UTF8

$startScript = @"
@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$ProjectsPath\byteport-services.ps1" -Action start
exit /b %ERRORLEVEL%
"@

$startScript | Out-File -FilePath "$ProjectsPath\start-services.bat" -Encoding ASCII

$stopScript = @"
@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$ProjectsPath\byteport-services.ps1" -Action stop
exit /b %ERRORLEVEL%
"@

$stopScript | Out-File -FilePath "$ProjectsPath\stop-services.bat" -Encoding ASCII

# Set up firewall rules
Write-Host "🔥 Configuring Windows Firewall..." -ForegroundColor Yellow

$ports = @(8081, 3000, 5173)
foreach ($port in $ports) {
    try {
        netsh advfirewall firewall add rule name="BytePort-$port" dir=in action=allow protocol=TCP localport=$port 2>$null
        Write-Host "Opened port $port" -ForegroundColor Gray
    } catch {
        Write-Host "Warning: Could not open port $port" -ForegroundColor Yellow
    }
}

# Create project template
Write-Host "📋 Creating project template..." -ForegroundColor Yellow

$projectTemplate = @"
NAME: "example-project"
DESCRIPTION: "Example project description"
SERVICES:
  - NAME: "main"
    PATH: "./frontend"
    PORT: 8080
  - NAME: "api"
    PATH: "./backend"
    PORT: 8081
"@

$projectTemplate | Out-File -FilePath "$ProjectsPath\odin.nvms.template" -Encoding UTF8

Write-Host "✅ BytePort Windows Server setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Your BytePort repository should be cloned to this directory" -ForegroundColor White
Write-Host "2. Configure Cloudflare tunnel (if not skipped)" -ForegroundColor White
Write-Host "3. Install frontend dependencies: cd frontend\web && npm install" -ForegroundColor White
Write-Host "4. Run start-services.bat to start all services" -ForegroundColor White
Write-Host ""
Write-Host "Configuration files created:" -ForegroundColor Yellow
Write-Host "- Environment: $ProjectsPath\.env" -ForegroundColor White
Write-Host "- Services: $ProjectsPath\start-services.bat" -ForegroundColor White
Write-Host "- Template: $ProjectsPath\odin.nvms.template" -ForegroundColor White
Write-Host ""
Write-Host "Access BytePort at: http://localhost:5173" -ForegroundColor Green
