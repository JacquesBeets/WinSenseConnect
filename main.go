package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"
	"unsafe"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kardianos/service"
	"golang.org/x/sys/windows"
)

var (
	modWtsapi32              = windows.NewLazySystemDLL("Wtsapi32.dll")
	procWTSQueryUserToken    = modWtsapi32.NewProc("WTSQueryUserToken")
	procWTSEnumerateSessions = modWtsapi32.NewProc("WTSEnumerateSessionsW")
	procWTSFreeMemory        = modWtsapi32.NewProc("WTSFreeMemory")
)

type WTS_SESSION_INFO struct {
	SessionID     uint32
	WindowStation *uint16
	State         uint32
}

type Config struct {
	BrokerAddress string                  `json:"broker_address"`
	Username      string                  `json:"username"`
	Password      string                  `json:"password"`
	ClientID      string                  `json:"client_id"`
	Topic         string                  `json:"topic"`
	LogLevel      string                  `json:"log_level"`
	ScriptTimeout int                     `json:"script_timeout"`
	Commands      map[string]ScriptConfig `json:"commands"`
}

type ScriptConfig struct {
	ScriptPath string `json:"script_path"`
	RunAsUser  bool   `json:"run_as_user"`
}

type program struct {
	mqttClient mqtt.Client
	config     Config
	logger     *Logger
	scriptDir  string
}

func newProgram() (*program, error) {
	p := &program{}
	var err error
	p.logger, err = NewLogger("MQTTPowershellService.log", "debug", "MQTTPowershellService")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	p.scriptDir = filepath.Join(filepath.Dir(exePath), "scripts")

	return p, nil
}

func (p *program) loadConfig() error {
	p.logger.Debug("Starting to load config...")
	exePath, err := os.Executable()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to get executable path: %v", err))
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	p.logger.Debug(fmt.Sprintf("Executable path: %s", exePath))

	configPath := filepath.Join(filepath.Dir(exePath), "config.json")
	p.logger.Debug(fmt.Sprintf("Config path: %s", configPath))

	file, err := os.Open(configPath)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to open config file: %v", err))
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&p.config); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to decode config: %v", err))
		return fmt.Errorf("failed to decode config: %v", err)
	}

	p.logger.Debug("Config loaded successfully")
	return nil
}

func (p *program) Start(s service.Service) error {
	p.logger.Debug("Starting service")
	if err := p.loadConfig(); err != nil {
		errMsg := fmt.Sprintf("Failed to load config: %v", err)
		p.logger.Error(errMsg)
		return err
	}
	p.logger.Debug("Config loaded, about to start run function")
	go p.run()
	return nil
}

func getActiveSessionID() (uint32, error) {
	var count uint32
	var sessions uintptr
	ret, _, err := procWTSEnumerateSessions.Call(0, 0, 1, uintptr(unsafe.Pointer(&sessions)), uintptr(unsafe.Pointer(&count)))
	if ret == 0 {
		return 0, fmt.Errorf("WTSEnumerateSessions failed: %v", err)
	}
	defer procWTSFreeMemory.Call(sessions)

	type WTS_SESSION_INFO struct {
		SessionID     uint32
		WindowStation *uint16
		State         uint32
	}

	for i := uint32(0); i < count; i++ {
		session := *(*WTS_SESSION_INFO)(unsafe.Pointer(sessions + uintptr(i)*unsafe.Sizeof(WTS_SESSION_INFO{})))
		if session.State == 0 { // WTSActive
			return session.SessionID, nil
		}
	}

	return 0, fmt.Errorf("no active session found")
}

func wtsQueryUserToken(session uint32, token *windows.Token) error {
	r1, _, e1 := procWTSQueryUserToken.Call(uintptr(session), uintptr(unsafe.Pointer(token)))
	if r1 == 0 {
		return error(e1)
	}
	return nil
}

func (p *program) run() {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error(fmt.Sprintf("Recovered from panic in run: %v\nStack trace: %s", r, debug.Stack()))
		}
	}()

	p.logger.Debug("Run function started")

	opts := mqtt.NewClientOptions().AddBroker(p.config.BrokerAddress)
	opts.SetClientID(p.config.ClientID)
	opts.SetUsername(p.config.Username)
	opts.SetPassword(p.config.Password)
	opts.SetOnConnectHandler(p.onConnect)
	opts.SetConnectionLostHandler(p.onConnectionLost)

	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Minute * 5)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(time.Second * 10)

	p.mqttClient = mqtt.NewClient(opts)

	for {
		p.logger.Debug(fmt.Sprintf("Attempting to connect to MQTT broker at %s...", p.config.BrokerAddress))
		if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
			p.logger.Error(fmt.Sprintf("Connection failed: %v", token.Error()))
			time.Sleep(time.Second * 10)
		} else {
			p.logger.Debug("Connection successful")
			break
		}
	}

	for {
		if !p.mqttClient.IsConnected() {
			p.logger.Debug("Connection lost, attempting to reconnect...")
			if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
				p.logger.Error(fmt.Sprintf("Reconnection failed: %v", token.Error()))
			} else {
				p.logger.Debug("Reconnection successful")
			}
		} else {
			p.logger.Debug("MQTT client is connected")
		}
		time.Sleep(time.Minute)
		p.logger.Debug("Service is still running...")
	}
}

func (p *program) onConnect(client mqtt.Client) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error(fmt.Sprintf("Recovered from panic in onConnect: %v\nStack trace: %s", r, debug.Stack()))
		}
	}()

	p.logger.Debug("Connected to MQTT broker")

	// Subscribe to the command topic
	if token := client.Subscribe(p.config.Topic, 0, p.commandHandler); token.Wait() && token.Error() != nil {
		errMsg := fmt.Sprintf("Failed to subscribe to command topic: %v", token.Error())
		p.logger.Error(errMsg)
	} else {
		p.logger.Debug(fmt.Sprintf("Successfully subscribed to command topic: %s", p.config.Topic))
	}

	// Subscribe to the response topic
	responseTopic := p.config.Topic + "/response"
	if token := client.Subscribe(responseTopic, 0, p.responseHandler); token.Wait() && token.Error() != nil {
		errMsg := fmt.Sprintf("Failed to subscribe to response topic: %v", token.Error())
		p.logger.Error(errMsg)
	} else {
		p.logger.Debug(fmt.Sprintf("Successfully subscribed to response topic: %s", responseTopic))
	}
}

func (p *program) onConnectionLost(client mqtt.Client, err error) {
	p.logger.Error(fmt.Sprintf("Connection to MQTT broker lost: %v", err))
}

func (p *program) runAsLoggedInUser(scriptPath string) (string, error) {
	sessionID, err := getActiveSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to get active session ID: %v", err)
	}

	var userToken windows.Token
	err = wtsQueryUserToken(sessionID, &userToken)
	if err != nil {
		return "", fmt.Errorf("failed to get user token: %v", err)
	}
	defer userToken.Close()

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command",
		fmt.Sprintf("Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope Process; & '%s'", scriptPath))
	cmd.SysProcAttr = &syscall.SysProcAttr{Token: syscall.Token(userToken)}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, output)
	}

	return string(output), nil
}

func (p *program) runAsLocalSystem(scriptPath string) (string, error) {
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
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

func (p *program) commandHandler(client mqtt.Client, msg mqtt.Message) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error(fmt.Sprintf("Recovered from panic in commandHandler: %v\nStack trace: %s", r, debug.Stack()))
		}
	}()

	command := string(msg.Payload())
	p.logger.Debug(fmt.Sprintf("Received command: %s", command))

	scriptConfig, exists := p.config.Commands[command]
	if !exists {
		p.logger.Error(fmt.Sprintf("Unknown command: %s", command))
		return
	}

	scriptPath := filepath.Join(p.scriptDir, scriptConfig.ScriptPath)
	p.logger.Debug(fmt.Sprintf("Executing script: %s", scriptPath))

	output, err := p.executeScript(scriptPath, scriptConfig.RunAsUser)
	if err != nil {
		errMsg := fmt.Sprintf("Error executing script for command '%s': %v", command, err)
		p.logger.Error(errMsg)
		p.publishResponse(client, errMsg)
	} else {
		p.logger.Debug(fmt.Sprintf("Successfully executed command: %s\nOutput: %s", command, output))
		p.publishResponse(client, output)
	}
}

func (p *program) responseHandler(client mqtt.Client, msg mqtt.Message) {
	p.logger.Debug(fmt.Sprintf("Received response: %s", string(msg.Payload())))
}

func (p *program) publishResponse(client mqtt.Client, message string) {
	responseTopic := p.config.Topic + "/response"
	if token := client.Publish(responseTopic, 0, false, message); token.Wait() && token.Error() != nil {
		p.logger.Error(fmt.Sprintf("Failed to publish script output: %v", token.Error()))
	}
}

func (p *program) Stop(s service.Service) error {
	p.logger.Debug("Stopping service")
	if p.mqttClient != nil && p.mqttClient.IsConnected() {
		p.mqttClient.Disconnect(250)
	}
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "MQTTPowershellService",
		DisplayName: "MQTT Powershell Automation Service",
		Description: "Listens for MQTT messages and runs PowerShell scripts",
	}

	prg, err := newProgram()
	if err != nil {
		fmt.Printf("Failed to create program: %v\n", err)
		return
	}
	defer prg.logger.Close()

	s, err := service.New(prg, svcConfig)
	if err != nil {
		prg.logger.Error(fmt.Sprintf("Failed to create service: %v", err))
		return
	}

	prg.logger.Debug("Service created, running service...")

	err = s.Run()
	if err != nil {
		prg.logger.Error(fmt.Sprintf("Service failed: %v", err))
		return
	}

	prg.logger.Debug("Service run completed")
}
