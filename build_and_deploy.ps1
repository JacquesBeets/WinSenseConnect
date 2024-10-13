# build_and_deploy.ps1

# Self-elevate the script if required
if (-Not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] 'Administrator')) {
 if ([int](Get-CimInstance -Class Win32_OperatingSystem | Select-Object -ExpandProperty BuildNumber) -ge 6000) {
  $CommandLine = "-File `"" + $MyInvocation.MyCommand.Path + "`" " + $MyInvocation.UnboundArguments
  Start-Process -FilePath PowerShell.exe -Verb Runas -ArgumentList $CommandLine
  Exit
 }
}

# Rest of the script starts here
Write-Host "Running with administrator privileges"

# Stop the service if it's running
Stop-Service -Name "WinSenseConnect" -ErrorAction SilentlyContinue

# Build the main Go program
Write-Host "Building the main Go program..."
Set-Location .\backend
$env:CGO_ENABLED=1; go build -o ..\WinSenseConnect.exe 

if ($LASTEXITCODE -ne 0) {
    Write-Host "Main program build failed. Exiting."
    exit 1
}

# Build the systray application
Write-Host "Building the systray application..."
Set-Location ..\systray
go build -o ..\WinSenseConnectSystray.exe

if ($LASTEXITCODE -ne 0) {
    Write-Host "Systray build failed. Exiting."
    exit 1
}

# Return to the root directory
Set-Location ..

# Remove the existing service
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

Start-Sleep -Seconds 2

# Start the service
Write-Host "Starting the service..."
Start-Service -Name "WinSenseConnect"

# Check the service status
$service = Get-Service -Name "WinSenseConnect"
Write-Host "Service status: $($service.Status)"

# Start the systray application
Write-Host "Starting the systray application..."
Start-Process -FilePath .\WinSenseConnectSystray.exe

Write-Host "Deployment complete!"

# Keep the window open
Read-Host -Prompt "Press Enter to exit"
