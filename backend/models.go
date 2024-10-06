package main

import (
	"time"
)

type ConfigModel struct {
	ID            int64     `db:"id"`
	BrokerAddress string    `db:"broker_address"`
	Username      string    `db:"username"`
	Password      string    `db:"password"`
	ClientID      string    `db:"client_id"`
	Topic         string    `db:"topic"`
	LogLevel      string    `db:"log_level"`
	ScriptTimeout int       `db:"script_timeout"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type ScriptConfig struct {
	ID            int64     `db:"id"`
	Name          string    `db:"name"`
	ScriptPath    string    `db:"script_path"`
	ScriptTimeout int       `db:"script_timeout"`
	RunAsUser     bool      `db:"run_as_user"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
type ScriptConfigs []ScriptConfig

type SensorConfig struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Enabled     bool      `db:"enabled"`
	Interval    int       `db:"interval"`
	SensorTopic string    `db:"sensor_topic"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
type SensorConfigs []SensorConfig
