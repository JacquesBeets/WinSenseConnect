# build_and_deploy.ps1

# Stop the service if it's running
Stop-Service -Name "MQTTPowershellService" -ErrorAction SilentlyContinue

# Build the Go program
Write-Host "Building the Go program..."
go build -o MQTTPowershellService.exe

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed. Exiting."
    exit 1
}

# Remove the existing service
Write-Host "Removing existing service..."
sc.exe delete MQTTPowershellService

# Remove the log file if it exists
$mqttLogPath = (Resolve-Path .\MQTTPowershellService.log).Path
if(Test-Path $mqttLogPath) {
    Remove-Item $mqttLogPath
}

Start-Sleep -Seconds 2

# Install the new service
Write-Host "Installing new service..."
$binaryPath = (Resolve-Path .\MQTTPowershellService.exe).Path

sc.exe create MQTTPowershellService binPath= "$binaryPath" start= auto obj= LocalSystem type= interact type= own DisplayName= "MQTT Powershell Automation Service"

# Set description and display name
sc.exe description MQTTPowershellService "Listens for MQTT messages and runs PowerShell scripts"
# sc.exe config MQTTPowershellService DisplayName= "MQTT Powershell Automation Service" type= interact type= own

# Set the required privilege
Write-Host "Setting required privileges..."
# sc.exe privs MQTTPowershellService SeAssignPrimaryTokenPrivilege

Start-Sleep -Seconds 2

# Start the service
Write-Host "Starting the service..."
Start-Service -Name "MQTTPowershellService"

# Check the service status
$service = Get-Service -Name "MQTTPowershellService"
Write-Host "Service status: $($service.Status)"

# Verify privileges
Write-Host "Verifying service privileges..."
$privs = sc.exe privs MQTTPowershellService
Write-Host $privs

Write-Host "Deployment complete!"