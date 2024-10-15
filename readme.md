# MQTT Windows Automation Service

This project is a Windows service that listens for MQTT messages and executes PowerShell scripts based on received commands. It's designed for home automation tasks, allowing remote control of a Windows PC through MQTT messages.

## Prerequisites

- Go 1.15 or later
- Windows 10 or later
- PowerShell 5.1 or later
- An MQTT broker (e.g., Mosquitto) set up and running

## Installation

1. Download the latest release of WinSenseConnect from the GitHub releases page.

2. Extract the downloaded zip file to a directory of your choice.

3. Open PowerShell as Administrator and run the following commands to install and start the service:
   - Replace `C:\path\to\extracted\folder` with the path where you extracted the files.
  
   ```powershell
   New-Service -Name "WinSenseConnect" -BinaryPathName "C:\path\to\extracted\folder\WinSenseConnect.exe" -DisplayName "MQTT Powershell Automation Service" -StartupType Automatic -Description "Listens for MQTT messages and runs PowerShell scripts"
   Start-Service -Name "WinSenseConnect"
   ```

4. After installation, open a web browser and navigate to `http://localhost:8080` to access the web dashboard.

5. Use the web dashboard to configure your MQTT settings, manage scripts, and view logs.

## Usage

Once the service is running and configured through the web dashboard, it will listen for messages on the specified MQTT topic. When a message is received, it will execute the corresponding PowerShell script.

To trigger a command, publish a message to your MQTT topic with the command as the payload. For example, to switch to your MacBook, you would publish the message "switch_to_macbook" to the topic you configured in the dashboard.

## Web Dashboard

The web dashboard provides an easy-to-use interface for managing your WinSenseConnect service. Here's what you can do:

1. Configure MQTT settings: Set your broker address, credentials, and topics.
2. Manage scripts: Add, edit, or remove PowerShell scripts that can be triggered via MQTT.
3. View logs: Check the service logs directly from the dashboard.
4. Monitor service status: See if the service is running and connected to the MQTT broker.

## Logging

The service logs its activities to two places:

1. Windows Event Log: You can view these logs in the Event Viewer under Windows Logs > Application.
2. Web Dashboard: Logs can be viewed directly in the web interface.

## Modifying Commands

To add or modify commands:

1. Open the web dashboard at `http://localhost:8077`.
2. Navigate to the Scripts section.
3. Add a new script or edit an existing one.
4. Save your changes.

The service will automatically reload the configuration, so there's no need to restart it.

## Troubleshooting

If you encounter issues:

1. Check the logs in the web dashboard.
2. Ensure your MQTT broker is running and accessible.
3. Verify that the PowerShell scripts exist and are correctly configured in the dashboard.
4. Check that the service is running:

   ```powershell
   Get-Service -Name "WinSenseConnect"
   ```

## Uninstalling

To remove the service:

1. Stop and delete the service:

   ```powershell
   Stop-Service -Name "WinSenseConnect"
   Remove-Service -Name "WinSenseConnect"
   ```

   For older versions of PowerShell:

   ```powershell
   sc.exe delete "WinSenseConnect"
   ```

2. Delete the WinSenseConnect folder and all its contents.

## Security Considerations

- Access to the web dashboard should be restricted to trusted users only.
- Be cautious about what commands you allow and what the PowerShell scripts do.
- Consider network-level security to restrict access to your MQTT broker and the web dashboard.
- The service uses a secure method to store sensitive information like MQTT credentials.

## Contributing

Contributions to improve the service are welcome. Please feel free to submit pull requests or open issues for bugs and feature requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
