package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kardianos/service"
)

type Config struct {
	BrokerAddress string            `json:"broker_address"`
	Username      string            `json:"username"`
	Password      string            `json:"password"`
	ClientID      string            `json:"client_id"`
	Topic         string            `json:"topic"`
	LogLevel      string            `json:"log_level"`
	Commands      map[string]string `json:"commands"`
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

	// Set the script directory
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

func (p *program) run() {
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

	// Attempt initial connection
	p.logger.Debug("Attempting initial connection to MQTT broker...")
	if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		p.logger.Error(fmt.Sprintf("Initial connection failed: %v", token.Error()))
	} else {
		p.logger.Debug("Initial connection successful")
	}

	// Keep the service running
	for {
		time.Sleep(time.Minute)
		p.logger.Debug("Service is still running...")
	}
}

func (p *program) onConnect(client mqtt.Client) {
	p.logger.Debug("Reconnected to MQTT broker")
	if token := client.Subscribe(p.config.Topic, 0, p.messageHandler); token.Wait() && token.Error() != nil {
		errMsg := fmt.Sprintf("Failed to subscribe to topic: %v", token.Error())
		p.logger.Error(errMsg)
	}
}

func (p *program) onConnectionLost(client mqtt.Client, err error) {
	errMsg := fmt.Sprintf("Connection to MQTT broker lost: %v", err)
	p.logger.Error(errMsg)
}

func (p *program) messageHandler(client mqtt.Client, msg mqtt.Message) {
	command := string(msg.Payload())
	p.logger.Debug(fmt.Sprintf("Received command: %s", command))

	scriptName, exists := p.config.Commands[command]
	if !exists {
		p.logger.Error(fmt.Sprintf("Unknown command: %s", command))
		return
	}

	scriptPath := filepath.Join(p.scriptDir, scriptName)

	p.logger.Debug(fmt.Sprintf("Executing script: %s", scriptPath))

	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)

	// Capture stdout and stderr
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error executing script for command '%s': %v\nStderr: %s", command, err, stderr.String()))
	} else {
		p.logger.Debug(fmt.Sprintf("Successfully executed command: %s\nOutput: %s", command, out.String()))
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
