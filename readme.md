# MQTT Windows Automation Service

This project is a Windows service that listens for MQTT messages and executes PowerShell scripts based on received commands. It's designed for home automation tasks, allowing remote control of a Windows PC through MQTT messages.

## Prerequisites

- Go 1.15 or later
- Windows 10 or later
- PowerShell 5.1 or later
- An MQTT broker (e.g., Mosquitto) set up and running

## Installation

1. Clone this repository or download the source code.

2. Install the required Go packages:
   ```
   go get github.com/eclipse/paho.mqtt.golang
   go get github.com/kardianos/service
   go get golang.org/x/sys/windows/svc/eventlog
   ```

3. Create a `config.json` file in the same directory as the Go script:
   ```json
   {
       "broker_address": "tcp://your_broker_ip:1883",
       "username": "your_mqtt_username",
       "password": "your_mqtt_password",
       "client_id": "windows-automation-service",
       "topic": "windows/commands",
       "commands": {
           "switch_to_macbook": "C:\\scripts\\switch_to_macbook.ps1",
           "shutdown_windows": "C:\\scripts\\shutdown_windows.ps1",
           "restart_windows": "C:\\scripts\\restart_windows.ps1",
           "launch_app": "C:\\scripts\\launch_app.ps1"
       }
   }
   ```
   Replace the broker address, username, and password with your MQTT broker details.

4. Create the PowerShell scripts referenced in your `config.json`. For example:
   - `C:\scripts\switch_to_macbook.ps1`
   - `C:\scripts\shutdown_windows.ps1`
   - `C:\scripts\restart_windows.ps1`
   - `C:\scripts\launch_app.ps1`

5. Build the Go program:
   ```
   go build -o MQTTPowershellService.exe
   ```

6. Open PowerShell as Administrator and run the following commands to install and start the service:
   ```powershell
   New-Service -Name "MQTTPowershellService" -BinaryPathName "C:\path\to\MQTTPowershellService.exe" -DisplayName "MQTT Powershell Automation Service" -StartupType Automatic -Description "Listens for MQTT messages and runs PowerShell scripts"
   Start-Service -Name "MQTTPowershellService"
   ```

## Usage

Once the service is running, it will listen for messages on the MQTT topic specified in your `config.json`. When a message is received, it will execute the corresponding PowerShell script.

To trigger a command, publish a message to your MQTT topic with the command as the payload. For example, to switch to your MacBook, you would publish the message "switch_to_macbook" to the topic "windows/commands" (or whatever topic you specified in your config).

## Logging

The service logs its activities to two places:

1. Windows Event Log: You can view these logs in the Event Viewer under Windows Logs > Application.
2. A log file: Located at `C:\MQTTPowershellService.log`.

## Modifying Commands

To add or modify commands:

1. Stop the service:
   ```powershell
   Stop-Service -Name "MQTTPowershellService"
   ```

2. Edit the `config.json` file to add or change command entries.

3. Create or modify the corresponding PowerShell scripts.

4. Start the service:
   ```powershell
   Start-Service -Name "MQTTPowershellService"
   ```

## Troubleshooting

If you encounter issues:

1. Check the log file at `C:\MQTTPowershellService.log`.
2. Ensure your MQTT broker is running and accessible.
3. Verify that the PowerShell scripts exist at the paths specified in `config.json`.
4. Check that the service is running:
   ```powershell
   Get-Service -Name "MQTTPowershellService"
   ```

## Uninstalling

To remove the service:

1. Stop and delete the service:
   ```powershell
   Stop-Service -Name "MQTTPowershellService"
   sc.exe delete "MQTTPowershellService"
   ```

2. Delete the executable and configuration files.

## Security Considerations

- Ensure your MQTT broker is properly secured.
- Use strong passwords for MQTT authentication.
- Be cautious about what commands you allow and what the PowerShell scripts do.
- Consider network-level security to restrict access to your MQTT broker.

## Contributing

Contributions to improve the service are welcome. Please feel free to submit pull requests or open issues for bugs and feature requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.