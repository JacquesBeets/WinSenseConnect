package main

import (
	"time"
)

type Config struct {
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

type Command struct {
	ID         int64     `db:"id"`
	Name       string    `db:"name"`
	ScriptPath string    `db:"script_path"`
	RunAsUser  bool      `db:"run_as_user"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type SensorConfig struct {
	ID          int64     `db:"id"`
	Enabled     bool      `db:"enabled"`
	Interval    int       `db:"interval"`
	SensorTopic string    `db:"sensor_topic"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
