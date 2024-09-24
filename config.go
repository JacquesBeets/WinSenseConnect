package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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
