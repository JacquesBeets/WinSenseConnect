# build_and_deploy.ps1

# Stop the service if it's running
Stop-Service -Name "WinSenseConnect" -ErrorAction SilentlyContinue

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
Write-Host "Service Removed!"
