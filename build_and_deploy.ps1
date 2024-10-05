# build_and_deploy.ps1

# Stop the service if it's running
Stop-Service -Name "WinSenseConnect" -ErrorAction SilentlyContinue

# Build the Go program
Write-Host "Building the Go program..."
Set-Location .\backend
go build -o ..\WinSenseConnect.exe

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed. Exiting."
    exit 1
}

# Remove the existing service
Set-Location ..
Write-Host "Removing existing service..."
sc.exe delete WinSenseConnect

# Remove the log file if it exists
$mqttLogPath = (Resolve-Path .\WinSenseConnect.log).Path
if(Test-Path $mqttLogPath) {
    Remove-Item $mqttLogPath
}

Start-Sleep -Seconds 2

# Install the new service
Write-Host "Installing new service..."
$binaryPath = (Resolve-Path .\WinSenseConnect.exe).Path

sc.exe create WinSenseConnect binPath= "$binaryPath" start= auto obj= LocalSystem type= interact type= own DisplayName= "WinSense MQTT & Server Service"

# Set description and display name
sc.exe description WinSenseConnect "Listens for MQTT messages and runs PowerShell scripts"
# sc.exe config WinSenseConnect DisplayName= "MQTT Powershell Automation Service" type= interact type= own

# Set the required privilege
Write-Host "Setting required privileges..."
# sc.exe privs WinSenseConnect SeAssignPrimaryTokenPrivilege

Start-Sleep -Seconds 2

# Start the service
Write-Host "Starting the service..."
Start-Service -Name "WinSenseConnect"

# Check the service status
$service = Get-Service -Name "WinSenseConnect"
Write-Host "Service status: $($service.Status)"

# Verify privileges
Write-Host "Verifying service privileges..."
$privs = sc.exe privs WinSenseConnect
Write-Host $privs

Write-Host "Deployment complete!"