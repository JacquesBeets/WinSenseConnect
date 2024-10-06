package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	db, err := sql.Open("sqlite3", "../data/store.db")
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) InitSchema() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			broker_address TEXT NOT NULL,
			username TEXT,
			password TEXT,
			client_id TEXT,
			topic TEXT,
			log_level TEXT,
			script_timeout INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		);

		CREATE TABLE IF NOT EXISTS commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			script_path TEXT NOT NULL,
			run_as_user BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME
		);

		CREATE TABLE IF NOT EXISTS sensor_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			enabled BOOLEAN,
			interval INTEGER,
			sensor_topic TEXT,
			created_at DATETIME,
			updated_at DATETIME
		);
	`)
	return err
}

func (db *DB) GetConfig() (*Config, error) {
	var config Config
	err := db.QueryRow("SELECT * FROM configs ORDER BY id DESC LIMIT 1").Scan(
		&config.ID, &config.BrokerAddress, &config.Username, &config.Password,
		&config.ClientID, &config.Topic, &config.LogLevel, &config.ScriptTimeout,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (db *DB) SaveConfig(config *Config) error {
	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO configs (
			broker_address, username, password, client_id, topic,
			log_level, script_timeout, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		config.BrokerAddress, config.Username, config.Password,
		config.ClientID, config.Topic, config.LogLevel,
		config.ScriptTimeout, now, now,
	)
	return err
}

func (db *DB) GetSensorConfig() (*SensorConfig, error) {
	var sensorConfig SensorConfig
	err := db.QueryRow("SELECT * FROM sensor_configs ORDER BY id DESC LIMIT 1").Scan(
		&sensorConfig.ID, &sensorConfig.Enabled, &sensorConfig.Interval,
		&sensorConfig.SensorTopic, &sensorConfig.CreatedAt, &sensorConfig.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sensorConfig, nil
}

func (db *DB) SaveSensorConfig(sensorConfig *SensorConfig) error {
	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO sensor_configs (
			enabled, interval, sensor_topic, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?)`,
		sensorConfig.Enabled, sensorConfig.Interval,
		sensorConfig.SensorTopic, now, now,
	)
	return err
}
