# build_systray.ps1

$env:CGO_ENABLED=1
go build -o ..\WinSenseConnectSystray.exe

if ($LASTEXITCODE -ne 0) {
    Write-Host "Systray build failed. Exiting."
    exit 1
}

Write-Host "Systray build completed successfully."
