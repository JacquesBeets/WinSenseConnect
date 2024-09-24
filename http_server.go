package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
)

func (p *program) startHTTPServer() {
	r := p.router

	// Serve static files (our UI) - this will be added at build time from our Nuxt frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui")))

	// API endpoints
	r.HandleFunc("/api/config", p.handleGetConfig).Methods("GET")
	r.HandleFunc("/api/config", p.handleUpdateConfig).Methods("POST")
	r.HandleFunc("/api/scripts", p.handleListScripts).Methods("GET")
	r.HandleFunc("/api/scripts", p.handleAddScript).Methods("POST")

	http.ListenAndServe(":8080", r)
}

func (p *program) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.config)
}

func (p *program) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig Config
	json.NewDecoder(r.Body).Decode(&newConfig)
	// Write new config to file
	// This will later be saved in db
}

func (p *program) handleListScripts(w http.ResponseWriter, r *http.Request) {
	files, _ := filepath.Glob(filepath.Join(p.scriptDir, "*.ps1"))
	json.NewEncoder(w).Encode(files)
}

func (p *program) handleAddScript(w http.ResponseWriter, r *http.Request) {
	// Logic to add new powershell scripts
}
