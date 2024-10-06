package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}

	dbpath := filepath.Join(filepath.Dir(exePath), "data", "store.db")
	_, err = os.Stat(dbpath)
	if os.IsNotExist(err) {
		// If it doesn't exist, create it
		_, err := os.Create(dbpath)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) InitSchema(logger *Logger) error {
	// Drop tables if they exist
	// logger.Debug("Dropping tables if they exist...")
	// _, err := db.Exec(`DROP TABLE IF EXISTS configs`)
	// if err != nil {
	// 	return err
	// }
	// _, err = db.Exec(`DROP TABLE IF EXISTS script_configs`)
	// if err != nil {
	// 	return err
	// }
	// _, err = db.Exec(`DROP TABLE IF EXISTS sensor_configs`)
	// if err != nil {
	// 	return err
	// }

	// Create tables
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
			script_timeout INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		);

		CREATE TABLE IF NOT EXISTS sensor_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			enabled BOOLEAN,
			interval INTEGER,
			sensor_topic TEXT,
			created_at DATETIME,
			updated_at DATETIME
		);
	`)

	if err != nil {
		return err
	}
	// Check if the default data already exists
	var defaultDataExists bool
	err = db.QueryRow("SELECT id FROM configs LIMIT 1").Scan(&defaultDataExists)
	logger.Debug(fmt.Sprintf("Default data exists: %v", defaultDataExists))
	if err != nil && err != sql.ErrNoRows {
		logger.Error(fmt.Sprintf("Failed to check if default data exists: %v", err))
		return err
	}
	if !defaultDataExists {
		// Add default data if it doesn't exist
		_, err = db.Exec(`
		INSERT INTO configs (id, broker_address, username, password, client_id, topic, log_level, script_timeout, created_at, updated_at)
		VALUES (1, 'tcp://0.0.0.0:1883', 'your_username', 'your_password', 'my-windows-automation-service', 'windows/commands', 'debug', 300, '2023-07-01 12:00:00', '2023-07-01 12:00:00');

		INSERT INTO script_configs (id, name, script_path, script_timeout, run_as_user, created_at, updated_at)
		VALUES (1, 'test_notification', 'test_notification.ps1', 300, true, '2023-07-01 12:00:00', '2023-07-01 12:00:00');

		INSERT INTO sensor_configs (id, name, enabled, interval, sensor_topic, created_at, updated_at)
		VALUES (1, 'cpu_usage', false, 60, 'windows/sensors/cpu_usage', '2023-07-01 12:00:00', '2023-07-01 12:00:00');
	`)
	}

	return err
}

func (db *DB) GetConfig() (*Config, error) {
	var configModel ConfigModel

	err := db.QueryRow("SELECT id, broker_address, username, password, client_id, topic, log_level, script_timeout, created_at, updated_at FROM configs ORDER BY id DESC LIMIT 1").Scan(
		&configModel.ID,
		&configModel.BrokerAddress,
		&configModel.Username,
		&configModel.Password,
		&configModel.ClientID,
		&configModel.Topic,
		&configModel.LogLevel,
		&configModel.ScriptTimeout,
		&configModel.CreatedAt,
		&configModel.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %v", err)
	}

	configsScript, err := db.GetScriptConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to get script configs: %v", err)
	}
	configsScriptArray := make(map[string]ScriptConfig)
	for _, config := range *configsScript {
		configsScriptArray[config.ScriptPath] = config
	}

	configsSensor, err := db.GetSensorConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to get sensor configs: %v", err)
	}
	configsSensorArray := make(map[string]SensorConfig)
	for _, config := range *configsSensor {
		configsSensorArray[config.SensorTopic] = config
	}

	config := Config{
		ID:                  configModel.ID,
		BrokerAddress:       configModel.BrokerAddress,
		Username:            configModel.Username,
		Password:            configModel.Password,
		ClientID:            configModel.ClientID,
		Topic:               configModel.Topic,
		LogLevel:            configModel.LogLevel,
		ScriptTimeout:       configModel.ScriptTimeout,
		SensorConfigEnabled: false,
		Commands:            configsScriptArray,
		Sensors:             configsSensorArray,
	}

	return &config, nil
}

func (db *DB) GetScriptConfigs() (*ScriptConfigs, error) {
	rows, err := db.Query("SELECT id, name, script_path, run_as_user, script_timeout, created_at, updated_at FROM script_configs ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query script configs: %v", err)
	}
	defer rows.Close()

	var scriptConfigs ScriptConfigs
	for rows.Next() {
		var sc ScriptConfig
		err := rows.Scan(&sc.ID, &sc.Name, &sc.ScriptPath, &sc.RunAsUser, &sc.ScriptTimeout, &sc.CreatedAt, &sc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan script config: %v", err)
		}
		scriptConfigs = append(scriptConfigs, sc)
	}
	return &scriptConfigs, nil
}

func (db *DB) GetSensorConfigs() (*SensorConfigs, error) {
	rows, err := db.Query("SELECT id, name, enabled, interval, sensor_topic, created_at, updated_at FROM sensor_configs ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query sensor configs: %v", err)
	}
	defer rows.Close()

	var sensorConfigs SensorConfigs
	for rows.Next() {
		var sc SensorConfig
		err := rows.Scan(&sc.ID, &sc.Name, &sc.Enabled, &sc.Interval, &sc.SensorTopic, &sc.CreatedAt, &sc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sensor config: %v", err)
		}
		sensorConfigs = append(sensorConfigs, sc)
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
	err := db.QueryRow("SELECT id, name, enabled, interval, sensor_topic, created_at, updated_at FROM sensor_configs WHERE id = ? ORDER BY id DESC LIMIT 1", id).Scan(
		&sensorConfig.ID,
		&sensorConfig.Name,
		&sensorConfig.Enabled,
		&sensorConfig.Interval,
		&sensorConfig.SensorTopic,
		&sensorConfig.CreatedAt,
		&sensorConfig.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get sensor config: %v", err)
	}
	return &sensorConfig, nil
}

func (db *DB) UpdateSensorConfig(sensorConfig *SensorConfig) error {
	_, err := db.Exec(`
		UPDATE sensor_configs SET
			name = ?, enabled = ?, interval = ?, sensor_topic = ?, updated_at = ?
		WHERE id = ?`,
		sensorConfig.Name,
		sensorConfig.Enabled,
		sensorConfig.Interval,
		sensorConfig.SensorTopic,
		time.Now(),
		sensorConfig.ID,
	)
	return err
}

func (db *DB) CreateSensorConfig(sensorConfig *SensorConfig) error {
	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO sensor_configs (
			name, enabled, interval, sensor_topic, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)`,
		sensorConfig.Name,
		sensorConfig.Enabled,
		sensorConfig.Interval,
		sensorConfig.SensorTopic,
		now,
		now,
	)
	return err
}
