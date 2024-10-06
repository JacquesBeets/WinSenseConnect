package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"github.com/kardianos/service"
	"golang.org/x/sys/windows"
)

type program struct {
	mqttClient mqtt.Client
	config     Config
	logger     *Logger
	scriptDir  string
	router     *mux.Router
	db         *DB
}

func newProgram() (*program, error) {
	p := &program{}
	var err error
	p.logger, err = NewLogger("WinSenseConnect.log", "debug", "WinSenseConnect")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}

	// Init DB
	p.db, err = NewDB()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create database: %v", err))
		return nil, err
	}

	// Init Schema
	err = p.db.InitSchema()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to initialize database schema: %v", err))
		return nil, err
	}

	// Set scripts directory
	p.scriptDir = filepath.Join(filepath.Dir(exePath), "scripts")

	// Init Router
	p.router = mux.NewRouter()

	return p, nil
}

func (p *program) Start(s service.Service) error {
	p.logger.Debug("Starting service")
	if err := p.loadConfig(); err != nil {
		errMsg := fmt.Sprintf("Failed to load config: %v", err)
		p.logger.Error(errMsg)
		return err
	}
	p.logger.Debug("Config loaded, about to start run function")
	go p.startHTTPServer()
	go p.run()
	return nil
}

func (p *program) run() {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error(fmt.Sprintf("Recovered from panic in run: %v\nStack trace: %s", r, debug.Stack()))
		}
	}()

	p.logger.Debug("Run function started")

	p.setupMQTTClient()

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

func (p *program) Stop(s service.Service) error {
	p.logger.Debug("Stopping service")
	if p.mqttClient != nil && p.mqttClient.IsConnected() {
		p.mqttClient.Disconnect(250)
	}
	return nil
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
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Token:         syscall.Token(userToken),
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
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
