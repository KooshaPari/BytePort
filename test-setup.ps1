# BytePort Windows Setup Test Script
# Verifies that all components are properly configured

param(
    [switch]$Verbose = $false
)

Write-Host "🧪 Testing BytePort Windows Setup..." -ForegroundColor Green
Write-Host ""

$testResults = @()
$allTestsPassed = $true

function Test-Component {
    param(
        [string]$Name,
        [scriptblock]$TestScript,
        [string]$SuccessMessage,
        [string]$FailureMessage
    )
    
    Write-Host "Testing $Name..." -ForegroundColor Yellow -NoNewline
    
    try {
        $result = & $TestScript
        if ($result) {
            Write-Host " ✅ PASS" -ForegroundColor Green
            $script:testResults += [PSCustomObject]@{
                Component = $Name
                Status = "PASS"
                Message = $SuccessMessage
            }
        } else {
            Write-Host " ❌ FAIL" -ForegroundColor Red
            $script:testResults += [PSCustomObject]@{
                Component = $Name
                Status = "FAIL"
                Message = $FailureMessage
            }
            $script:allTestsPassed = $false
        }
    } catch {
        Write-Host " ❌ ERROR" -ForegroundColor Red
        $script:testResults += [PSCustomObject]@{
            Component = $Name
            Status = "ERROR"
            Message = "Exception: $($_.Exception.Message)"
        }
        $script:allTestsPassed = $false
    }
}

# Test 1: Directory Structure
Test-Component -Name "Directory Structure" -TestScript {
    $requiredDirs = @(
        "C:\BytePort\projects",
        "C:\BytePort\tunnels",
        "C:\BytePort\logs",
        "C:\BytePort\backups"
    )
    
    foreach ($dir in $requiredDirs) {
        if (!(Test-Path $dir)) {
            if ($Verbose) { Write-Host "Missing directory: $dir" }
            return $false
        }
    }
    return $true
} -SuccessMessage "All required directories exist" -FailureMessage "Some required directories are missing"

# Test 2: Go Installation
Test-Component -Name "Go Installation" -TestScript {
    try {
        $goVersion = go version 2>$null
        return $goVersion -like "*go version*"
    } catch {
        return $false
    }
} -SuccessMessage "Go is installed and accessible" -FailureMessage "Go is not installed or not in PATH"

# Test 3: Node.js Installation
Test-Component -Name "Node.js Installation" -TestScript {
    try {
        $nodeVersion = node --version 2>$null
        return $nodeVersion -like "v*"
    } catch {
        return $false
    }
} -SuccessMessage "Node.js is installed and accessible" -FailureMessage "Node.js is not installed or not in PATH"

# Test 4: Docker Installation
Test-Component -Name "Docker Installation" -TestScript {
    try {
        $dockerVersion = docker version 2>$null
        return $dockerVersion -like "*Docker version*"
    } catch {
        return $false
    }
} -SuccessMessage "Docker is installed and running" -FailureMessage "Docker is not installed or not running"

# Test 5: Docker Network
Test-Component -Name "Docker Network" -TestScript {
    try {
        $networks = docker network ls 2>$null
        return $networks -like "*byteport-network*"
    } catch {
        return $false
    }
} -SuccessMessage "BytePort Docker network exists" -FailureMessage "BytePort Docker network not found"

# Test 6: Cloudflared Installation
Test-Component -Name "Cloudflared Installation" -TestScript {
    try {
        $cloudflaredVersion = cloudflared version 2>$null
        return $cloudflaredVersion -like "*cloudflared version*"
    } catch {
        return $false
    }
} -SuccessMessage "Cloudflared is installed and accessible" -FailureMessage "Cloudflared is not installed or not in PATH"

# Test 7: Environment Variables
Test-Component -Name "Environment Variables" -TestScript {
    $requiredVars = @(
        "BYTEPORT_ROOT",
        "BYTEPORT_DOMAIN",
        "PROJECTS_PATH",
        "TUNNEL_CONFIG_PATH"
    )
    
    foreach ($var in $requiredVars) {
        $value = [Environment]::GetEnvironmentVariable($var, "Machine")
        if ([string]::IsNullOrEmpty($value)) {
            if ($Verbose) { Write-Host "Missing environment variable: $var" }
            return $false
        }
    }
    return $true
} -SuccessMessage "All required environment variables are set" -FailureMessage "Some required environment variables are missing"

# Test 8: BytePort Backend Dependencies
Test-Component -Name "NVMS Backend Dependencies" -TestScript {
    $nvmsPath = ".\backend\nvms"
    if (Test-Path $nvmsPath) {
        Push-Location $nvmsPath
        try {
            $goModCheck = go mod verify 2>$null
            return $LASTEXITCODE -eq 0
        } finally {
            Pop-Location
        }
    }
    return $false
} -SuccessMessage "NVMS backend dependencies are valid" -FailureMessage "NVMS backend dependencies have issues"

# Test 9: BytePort API Dependencies
Test-Component -Name "API Backend Dependencies" -TestScript {
    $apiPath = ".\backend\byteport"
    if (Test-Path $apiPath) {
        Push-Location $apiPath
        try {
            $goModCheck = go mod verify 2>$null
            return $LASTEXITCODE -eq 0
        } finally {
            Pop-Location
        }
    }
    return $false
} -SuccessMessage "API backend dependencies are valid" -FailureMessage "API backend dependencies have issues"

# Test 10: Frontend Dependencies
Test-Component -Name "Frontend Dependencies" -TestScript {
    $frontendPath = ".\frontend\web"
    if (Test-Path $frontendPath) {
        $nodeModulesPath = Join-Path $frontendPath "node_modules"
        return Test-Path $nodeModulesPath
    }
    return $false
} -SuccessMessage "Frontend dependencies are installed" -FailureMessage "Frontend dependencies are missing"

# Test 11: Configuration Files
Test-Component -Name "Configuration Files" -TestScript {
    $configFiles = @(
        ".\setup-windows.ps1",
        ".\start-services.bat",
        ".\stop-services.bat",
        "C:\BytePort\.env"
    )
    
    foreach ($file in $configFiles) {
        if (!(Test-Path $file)) {
            if ($Verbose) { Write-Host "Missing configuration file: $file" }
            return $false
        }
    }
    return $true
} -SuccessMessage "All configuration files exist" -FailureMessage "Some configuration files are missing"

# Test 12: Firewall Rules
Test-Component -Name "Firewall Rules" -TestScript {
    $ports = @(8081, 3000, 5173)
    foreach ($port in $ports) {
        $rule = netsh advfirewall firewall show rule name="BytePort-$port" 2>$null
        if ($rule -notlike "*BytePort-$port*") {
            if ($Verbose) { Write-Host "Missing firewall rule for port: $port" }
            return $false
        }
    }
    return $true
} -SuccessMessage "All required firewall rules exist" -FailureMessage "Some firewall rules are missing"

# Display Results
Write-Host ""
Write-Host "📊 Test Results Summary:" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan

$passCount = ($testResults | Where-Object { $_.Status -eq "PASS" }).Count
$failCount = ($testResults | Where-Object { $_.Status -eq "FAIL" }).Count
$errorCount = ($testResults | Where-Object { $_.Status -eq "ERROR" }).Count
$totalCount = $testResults.Count

Write-Host "Total Tests: $totalCount" -ForegroundColor White
Write-Host "Passed: $passCount" -ForegroundColor Green
Write-Host "Failed: $failCount" -ForegroundColor Red
Write-Host "Errors: $errorCount" -ForegroundColor Yellow

if ($Verbose -or !$allTestsPassed) {
    Write-Host ""
    Write-Host "Detailed Results:" -ForegroundColor Cyan
    Write-Host "-----------------" -ForegroundColor Cyan
    
    foreach ($result in $testResults) {
        $color = switch ($result.Status) {
            "PASS" { "Green" }
            "FAIL" { "Red" }
            "ERROR" { "Yellow" }
        }
        Write-Host "$($result.Component): $($result.Status)" -ForegroundColor $color
        if ($Verbose) {
            Write-Host "  $($result.Message)" -ForegroundColor Gray
        }
    }
}

Write-Host ""

if ($allTestsPassed) {
    Write-Host "🎉 All tests passed! BytePort Windows setup is ready." -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "1. Configure Cloudflare tunnel credentials" -ForegroundColor White
    Write-Host "2. Run .\start-services.bat to start BytePort" -ForegroundColor White
    Write-Host "3. Access BytePort at http://localhost:5173" -ForegroundColor White
    Write-Host "4. Deploy your first project!" -ForegroundColor White
} else {
    Write-Host "❌ Some tests failed. Please review the issues above." -ForegroundColor Red
    Write-Host ""
    Write-Host "Common fixes:" -ForegroundColor Yellow
    Write-Host "- Run setup-windows.ps1 as Administrator" -ForegroundColor White
    Write-Host "- Restart Docker Desktop" -ForegroundColor White
    Write-Host "- Check internet connectivity" -ForegroundColor White
    Write-Host "- Verify all software is installed" -ForegroundColor White
}

Write-Host ""
Write-Host "For detailed setup instructions, see README-Windows.md" -ForegroundColor Cyan
