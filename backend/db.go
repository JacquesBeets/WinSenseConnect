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

		CREATE TABLE IF NOT EXISTS script_configs (
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
	var (
		configModel ConfigModel
		config      Config
	)

	err := db.QueryRow("SELECT * FROM configs ORDER BY id DESC LIMIT 1").Scan(&configModel)
	if err != nil {
		return nil, err
	}

	configsScript, err := db.GetScriptConfigs()
	if err != nil {
		return nil, err
	}
	configsScriptArray := make(map[string]ScriptConfig)
	for _, config := range *configsScript {
		configsScriptArray[config.ScriptPath] = config
	}

	configsSensor, err := db.GetSensorConfigs()
	if err != nil {
		return nil, err
	}
	configsSensorArray := make(map[string]SensorConfig)
	for _, config := range *configsSensor {
		configsSensorArray[config.SensorTopic] = config
	}

	config = Config{
		ID:                  configModel.ID,
		BrokerAddress:       configModel.BrokerAddress,
		Username:            configModel.Username,
		Password:            configModel.Password,
		ClientID:            configModel.ClientID,
		Topic:               configModel.Topic,
		LogLevel:            configModel.LogLevel,
		ScriptTimeout:       300,
		SensorConfigEnabled: false,
		Commands:            configsScriptArray,
		Sensors:             configsSensorArray,
	}

	return &config, nil
}

func (db *DB) GetScriptConfigs() (*ScriptConfigs, error) {
	var scriptConfigs ScriptConfigs
	err := db.QueryRow("SELECT * FROM script_configs ORDER BY id DESC").Scan(&scriptConfigs)
	if err != nil {
		return nil, err
	}
	return &scriptConfigs, nil
}

func (db *DB) GetSensorConfigs() (*SensorConfigs, error) {
	var sensorConfigs SensorConfigs
	err := db.QueryRow("SELECT * FROM sensor_configs ORDER BY id DESC").Scan(&sensorConfigs)
	if err != nil {
		return nil, err
	}
	return &sensorConfigs, nil
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

func (db *DB) UpdateConfig(config *Config) error {
	now := time.Now()
	_, err := db.Exec(`
		UPDATE configs SET
			broker_address = ?, username = ?, password = ?, client_id = ?, topic = ?,
			log_level = ?, script_timeout = ?, updated_at = ?
		WHERE id = ?`,
		config.BrokerAddress, config.Username, config.Password,
		config.ClientID, config.Topic, config.LogLevel,
		config.ScriptTimeout, now,
		config.ID,
	)
	return err
}

func (db *DB) GetSensorConfig(id int64) (*SensorConfig, error) {
	var sensorConfig SensorConfig
	err := db.QueryRow("SELECT * FROM sensor_configs WHERE id = ? ORDER BY id DESC LIMIT 1", id).Scan(
		&sensorConfig.ID, &sensorConfig.Enabled, &sensorConfig.Interval,
		&sensorConfig.SensorTopic, &sensorConfig.CreatedAt, &sensorConfig.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sensorConfig, nil
}

func (db *DB) UpdateSensorConfig(sensorConfig *SensorConfig) error {
	_, err := db.Exec(`
		UPDATE sensor_configs SET
			enabled = ?, interval = ?, sensor_topic = ?, updated_at = ?
		WHERE id = ?`,
		sensorConfig.Enabled, sensorConfig.Interval,
		sensorConfig.SensorTopic, time.Now(),
		sensorConfig.ID,
	)
	return err
}

func (db *DB) CreateSensorConfig(sensorConfig *SensorConfig) error {
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
