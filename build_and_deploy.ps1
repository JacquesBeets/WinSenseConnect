
# Self-elevate the script if required
if (-Not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] 'Administrator')) {
    if ([int](Get-CimInstance -Class Win32_OperatingSystem | Select-Object -ExpandProperty BuildNumber) -ge 6000) {
        $CommandLine = "-File `"" + $MyInvocation.MyCommand.Path + "`" " + $MyInvocation.UnboundArguments
        Start-Process -FilePath PowerShell.exe -Verb Runas -ArgumentList $CommandLine -Wait
        Exit
    }
}

Write-Host "Running with administrator privileges"

# Get the script's directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $scriptDir
Write-Host "Current directory: $scriptDir"

# Function to forcefully delete a file
function Force-Delete($path) {
    if (Test-Path $path) {
        Write-Host "Attempting to delete: $path"
        try {
            Remove-Item -Path $path -Force -ErrorAction Stop
            Write-Host "Successfully deleted: $path"
        }
        catch {
            Write-Host "Failed to delete $path. Error: $_"
            Write-Host "Attempting to release file handles..."
            try {
                $handle = [System.Diagnostics.Process]::GetCurrentProcess().Handle
                $null = [PSClrInterop]::ReleaseFile($handle, $path)
                Remove-Item -Path $path -Force -ErrorAction Stop
                Write-Host "Successfully deleted after releasing handles: $path"
            }
            catch {
                Write-Host "Still unable to delete $path. Error: $_"
            }
        }
    }
    else {
        Write-Host "File not found: $path"
    }
}

# Stop the service if it's running
Write-Host "Stopping WinSenseConnect service..."
Stop-Service -Name "WinSenseConnect" -Force -ErrorAction SilentlyContinue
Write-Host "Stopping WinSenseConnectSystray process..."
Stop-Process -Name "WinSenseConnectSystray" -Force -ErrorAction SilentlyContinue

# Wait for processes to fully stop
Start-Sleep -Seconds 5

# Remove the log file and delete the executables
$mqttLogPath = Join-Path $scriptDir "WinSenseConnect.log"
$systrayLogPath = Join-Path $scriptDir "WinSenseConnectSystray.log"
$winsenseConnectPath = Join-Path $scriptDir "WinSenseConnect.exe"
$winsenseConnectSystrayPath = Join-Path $scriptDir "WinSenseConnectSystray.exe"

Write-Host "Attempting to delete files:"
Write-Host "Log files: $mqttLogPath, $systrayLogPath"
Write-Host "WinSenseConnect: $winsenseConnectPath"
Write-Host "WinSenseConnectSystray: $winsenseConnectSystrayPath"

Force-Delete $mqttLogPath
Force-Delete $systrayLogPath
Force-Delete $winsenseConnectPath
Force-Delete $winsenseConnectSystrayPath

# Additional wait to ensure files are released
Start-Sleep -Seconds 2

# Build the main service
Write-Host "Building the main service..."
$serviceDir = Join-Path $scriptDir "cmd\service"
Set-Location $serviceDir
$env:CGO_ENABLED=1; go build -o (Join-Path $scriptDir "WinSenseConnect.exe")

if ($LASTEXITCODE -ne 0) {
    Write-Host "Main program build failed. Exiting."
    exit 1
}

Start-Sleep -Seconds 2

# Build the systray application
Write-Host "Building the systray application..."
$systrayDir = Join-Path $scriptDir "cmd\systray"
Set-Location $systrayDir
$env:CGO_ENABLED=1
go build -o (Join-Path $scriptDir "WinSenseConnectSystray.exe") -ldflags "-H=windowsgui"
# go build -o (Join-Path $scriptDir "WinSenseConnectSystray.exe") // Development

if ($LASTEXITCODE -ne 0) {
    Write-Host "Systray build failed. Exiting."
    exit 1
}

# Return to the script directory
Set-Location $scriptDir

# Remove the existing service
Write-Host "Removing existing service..."
sc.exe delete WinSenseConnect

Start-Sleep -Seconds 2

# Install the new service
Write-Host "Installing new service..."
$binaryPath = (Resolve-Path (Join-Path $scriptDir "WinSenseConnect.exe")).Path

sc.exe create WinSenseConnect binPath= "$binaryPath" start= auto obj= LocalSystem type= own DisplayName= "WinSense MQTT & Server Service"

# Set description and display name
sc.exe description WinSenseConnect "Listens for MQTT messages and runs PowerShell scripts"

Start-Sleep -Seconds 2

# Start the service
Write-Host "Starting the service..."
Start-Service -Name "WinSenseConnect"

# Check the service status
$service = Get-Service -Name "WinSenseConnect"
Write-Host "Service status: $($service.Status)"

Start-Sleep -Seconds 2

# Start the systray application
Write-Host "Starting the systray application..."
$binaryPathSystray = (Resolve-Path (Join-Path $scriptDir "WinSenseConnectSystray.exe")).Path
Start-Process -FilePath $binaryPathSystray

Write-Host "Deployment complete!"

# Pause to keep the window open
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
