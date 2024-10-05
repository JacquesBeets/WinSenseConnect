package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) GetConfig() (*Config, error) {
	// Implement fetching config from database
	return nil, nil
}

func (db *DB) SaveConfig(config *Config) error {
	// Implement saving config to database
	return nil
}
