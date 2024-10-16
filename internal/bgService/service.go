package bgService

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"win-sense-connect/internal/common"
	"win-sense-connect/internal/shared"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"github.com/kardianos/service"
	"golang.org/x/sys/windows"
)

type program struct {
	mqttClient    mqtt.Client
	config        common.Config
	Logger        common.Logger
	scriptDir     string
	router        *mux.Router
	db            *shared.DB
	eventChannels []chan []byte
	eventMutex    sync.Mutex
}

func NewProgram() (*program, error) {
	p := &program{
		eventChannels: make([]chan []byte, 0),
	}
	var err error

	// Create a temporary logger
	tempLogger, err := NewLogger("WinSenseConnect.log", nil, "WinSenseConnect", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary logger: %v", err)
	}

	// Init DB
	p.db, err = shared.NewDB()
	if err != nil {
		tempLogger.Error(fmt.Sprintf("Failed to create database: %v", err))
		return nil, err
	}

	// Load config
	if err := p.loadConfig(tempLogger); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Initialize final logger with loaded config
	p.Logger, err = NewLogger("WinSenseConnect.log", &p.config, "WinSenseConnect", p.broadcastEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	// Init Schema
	err = p.db.InitSchema(p.Logger)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to initialize database schema: %v", err))
		return nil, err
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}

	// Set scripts directory
	p.scriptDir = filepath.Join(filepath.Dir(exePath), "scripts")

	// Init Router
	p.router = mux.NewRouter()

	return p, nil
}

func (p *program) broadcastEvent(event []byte) {
	p.eventMutex.Lock()
	defer p.eventMutex.Unlock()

	for _, ch := range p.eventChannels {
		select {
		case ch <- event:
		default:
			// If the channel is full, we skip this client
		}
	}
}

func (p *program) Start(s service.Service) error {
	p.Logger.Debug("Starting service")
	p.Logger.Debug("Config loaded, about to start run function")
	go p.startHTTPServer()
	go p.run()

	// Start systray if it's not running
	if err := p.startSystrayIfNotRunning(); err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to start systray: %v", err))
	}

	return nil
}

func (p *program) run() {
	defer func() {
		if r := recover(); r != nil {
			p.Logger.Error(fmt.Sprintf("Recovered from panic in run: %v\nStack trace: %s", r, debug.Stack()))
		}
	}()

	p.Logger.Debug("Run function started")

	p.setupMQTTClient()

	for {
		p.Logger.Debug(fmt.Sprintf("Attempting to connect to MQTT broker at %s...", p.config.BrokerAddress))
		if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
			p.Logger.Error(fmt.Sprintf("Connection failed: %v", token.Error()))
			time.Sleep(time.Second * 10)
		} else {
			p.Logger.Debug("Connection successful")
			break
		}
	}

	for {
		if !p.mqttClient.IsConnected() {
			p.Logger.Debug("Connection lost, attempting to reconnect...")
			if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
				p.Logger.Error(fmt.Sprintf("Reconnection failed: %v", token.Error()))
			} else {
				p.Logger.Debug("Reconnection successful")
			}
		} else {
			p.Logger.Debug("MQTT client is connected")
		}
		time.Sleep(time.Minute)
		p.Logger.Debug("Service is still running...")
	}
}

func (p *program) Stop(s service.Service) error {
	p.Logger.Debug("Stopping service")
	if p.mqttClient != nil && p.mqttClient.IsConnected() {
		p.mqttClient.Disconnect(250)
	}
	return nil
}

func (p *program) runAsLoggedInUser(scriptPath string) (string, error) {
	sessionID, err := getActiveSessionID()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to get active session ID: %v", err))
		return "", fmt.Errorf("failed to get active session ID: %v", err)
	}

	var userToken windows.Token
	err = wtsQueryUserToken(sessionID, &userToken)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to get user token: %v", err))
		return "", fmt.Errorf("failed to get user token: %v", err)
	}
	defer userToken.Close()

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command",
		fmt.Sprintf("Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope Process; & '%s'", scriptPath))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Token:         syscall.Token(userToken),
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("command failed: %v\nOutput: %s", err, output))
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, output)
	}

	return string(output), nil
}

func (p *program) runAsLocalSystem(scriptPath string) (string, error) {
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("command failed: %v\nOutput: %s", err, output))
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, output)
	}

	return string(output), nil
}

func (p *program) executeScript(scriptPath string, runAsUser bool) (string, error) {
	if runAsUser {
		return p.runAsLoggedInUser(scriptPath)
	} else {
		return p.runAsLocalSystem(scriptPath)
	}
}

func (p *program) restartService() error {
	p.Logger.Debug("Restarting service")
	err := p.Stop(nil)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to stop service: %v", err))
		return fmt.Errorf("failed to stop service: %v", err)
	}
	// clear mqtt client
	p.mqttClient = nil
	time.Sleep(time.Second * 5)
	p.Logger.Debug("Service stopped, restarting...")
	err = p.Start(nil)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to stop service: %v", err))
		return fmt.Errorf("failed to start service: %v", err)
	}
	return nil
}

func (p *program) isSystrayRunning() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq WinSenseConnectSystray.exe", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to check if systray is running: %v", err))
		return false
	}
	return len(output) > 0
}

func (p *program) startSystrayIfNotRunning() error {
	if !p.isSystrayRunning() {
		p.Logger.Debug("Systray is not running. Starting it now.")
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %v", err)
		}
		systrayPath := filepath.Join(filepath.Dir(exePath), "WinSenseConnectSystray.exe")
		cmd := exec.Command(systrayPath)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: windows.CREATE_NO_WINDOW,
		}
		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start systray: %v", err)
		}
		p.Logger.Debug("Systray started successfully.")
	} else {
		p.Logger.Debug("Systray is already running.")
	}
	return nil
}
