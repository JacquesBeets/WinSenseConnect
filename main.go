package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kardianos/service"
	"golang.org/x/sys/windows/svc/eventlog"
)

type Config struct {
	BrokerAddress string            `json:"broker_address"`
	Username      string            `json:"username"`
	Password      string            `json:"password"`
	ClientID      string            `json:"client_id"`
	Topic         string            `json:"topic"`
	Commands      map[string]string `json:"commands"`
}

type program struct {
	mqttClient mqtt.Client
	config     Config
	elog       *eventlog.Log
}

func (p *program) loadConfig() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	configPath := filepath.Join(filepath.Dir(exePath), "config.json")
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&p.config); err != nil {
		return fmt.Errorf("failed to decode config: %v", err)
	}

	return nil
}

func (p *program) logToFile(message string) {
	f, err := os.OpenFile("C:\\MQTTPowershellService.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	logger.Println(message)
}

func (p *program) Start(s service.Service) error {
	p.elog.Info(1, "Starting service")
	p.logToFile("Starting service")
	if err := p.loadConfig(); err != nil {
		errMsg := fmt.Sprintf("Failed to load config: %v", err)
		p.elog.Error(1, errMsg)
		p.logToFile(errMsg)
		return err
	}
	go p.run()
	return nil
}

func (p *program) run() {
	opts := mqtt.NewClientOptions().AddBroker(p.config.BrokerAddress)
	opts.SetClientID(p.config.ClientID)
	opts.SetUsername(p.config.Username)
	opts.SetPassword(p.config.Password)
	opts.SetOnConnectHandler(p.onConnect)
	opts.SetConnectionLostHandler(p.onConnectionLost)

	p.mqttClient = mqtt.NewClient(opts)
	if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		p.elog.Error(1, fmt.Sprintf("Failed to connect to MQTT broker: %v", token.Error()))
		return
	}

	p.elog.Info(1, "Connected to MQTT broker")
	p.elog.Info(1, "Connected to MQTT broker")
	p.logToFile("Connected to MQTT broker")

	for {
		time.Sleep(time.Second)
	}
}

func (p *program) onConnect(client mqtt.Client) {
	p.elog.Info(1, "Reconnected to MQTT broker")
	p.logToFile("Reconnected to MQTT broker")
	if token := client.Subscribe(p.config.Topic, 0, p.messageHandler); token.Wait() && token.Error() != nil {
		errMsg := fmt.Sprintf("Failed to subscribe to topic: %v", token.Error())
		p.elog.Error(1, errMsg)
		p.logToFile(errMsg)
	}
}

func (p *program) onConnectionLost(client mqtt.Client, err error) {
	errMsg := fmt.Sprintf("Connection to MQTT broker lost: %v", err)
	p.elog.Error(1, errMsg)
	p.logToFile(errMsg)
}

func (p *program) messageHandler(client mqtt.Client, msg mqtt.Message) {
	command := string(msg.Payload())
	logMsg := fmt.Sprintf("Received command: %s", command)
	p.elog.Info(1, logMsg)
	p.logToFile(logMsg)

	scriptPath, exists := p.config.Commands[command]
	if !exists {
		warnMsg := fmt.Sprintf("Unknown command: %s", command)
		p.elog.Warning(1, warnMsg)
		p.logToFile(warnMsg)
		return
	}

	cmd := exec.Command("powershell", "-File", scriptPath)
	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("Error executing script for command '%s': %v", command, err)
		p.elog.Error(1, errMsg)
		p.logToFile(errMsg)
	} else {
		successMsg := fmt.Sprintf("Successfully executed command: %s", command)
		p.elog.Info(1, successMsg)
		p.logToFile(successMsg)
	}
}

func (p *program) Stop(s service.Service) error {
	p.elog.Info(1, "Stopping service")
	p.logToFile("Stopping service")
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

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Printf("Failed to create service: %v\n", err)
		return
	}

	elog, err := eventlog.Open(svcConfig.Name)
	if err != nil {
		fmt.Printf("Failed to open event log: %v\n", err)
		return
	}
	defer elog.Close()

	prg.elog = elog

	if len(os.Args) > 1 {
		err = service.Control(s, os.Args[1])
		if err != nil {
			elog.Error(1, fmt.Sprintf("Failed to control service: %v", err))
			return
		}
		return
	}

	err = s.Run()
	if err != nil {
		elog.Error(1, fmt.Sprintf("Service failed: %v", err))
		return
	}
}
