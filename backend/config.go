package main

import (
	"fmt"
)

type Config struct {
	ID                  int64                   `json:"id"`
	BrokerAddress       string                  `json:"broker_address"`
	Username            string                  `json:"username"`
	Password            string                  `json:"password"`
	ClientID            string                  `json:"client_id"`
	Topic               string                  `json:"topic"`
	LogLevel            string                  `json:"log_level"`
	ScriptTimeout       int                     `json:"script_timeout"`
	SensorConfigEnabled bool                    `json:"sensor_config_enabled"`
	Commands            map[string]ScriptConfig `json:"commands"`
	Sensors             map[string]SensorConfig `json:"sensors"`
}

func (p *program) loadConfig(logger *Logger) error {
	logger.Debug("Starting to load config...")
	conf, err := p.db.GetConfig()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get config: %v", err))
		return err
	}
	p.config = *conf
	logger.Debug("Config loaded successfully")
	return nil
}
