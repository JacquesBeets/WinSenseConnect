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

# Install the new service
Write-Host "Installing new service..."
New-Service -Name "MQTTPowershellService" -BinaryPathName (Resolve-Path .\MQTTPowershellService.exe).Path -DisplayName "MQTT Powershell Automation Service" -StartupType Automatic

# Start the service
Write-Host "Starting the service..."
Start-Service -Name "MQTTPowershellService"

# Check the service status
$service = Get-Service -Name "MQTTPowershellService"
Write-Host "Service status: $($service.Status)"

Write-Host "Deployment complete!"