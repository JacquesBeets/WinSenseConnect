

# test-notification.ps1

# Display a notification using PowerShell's built-in cmdlets
$title = "Automation Server Test"
$message = "This is a test notification from your Go MQTT server!"

Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$balloon = New-Object System.Windows.Forms.NotifyIcon
$balloon.Icon = [System.Drawing.SystemIcons]::Information
$balloon.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
$balloon.BalloonTipTitle = $title
$balloon.BalloonTipText = $message
$balloon.Visible = $true

# Display the balloon tip
$balloon.ShowBalloonTip(5000)

# Also write to a log file for persistent evidence
$logPath = "C:\temp\automation_test_log.txt"
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Add-Content -Path $logPath -Value "$timestamp - Test notification displayed"

# Wait a moment before cleaning up
Start-Sleep -Seconds 6
$balloon.Dispose()

Write-Host "Notification displayed successfully."