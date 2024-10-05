package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/cors"
)

func (p *program) startHTTPServer() {
	p.logger.Debug("Starting HTTP server")
	r := p.router

	exePath, err := os.Executable()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to get executable path: %v", err))
	}
	staticPath := filepath.Join(filepath.Dir(exePath), "frontend/.output/public")

	// API endpoints
	r.HandleFunc("/api/config", p.handleGetConfig).Methods("GET")
	r.HandleFunc("/api/config", p.handleUpdateConfig).Methods("POST")
	r.HandleFunc("/api/scripts", p.handleListScripts).Methods("GET")
	r.HandleFunc("/api/scripts", p.handleAddScript).Methods("POST")
	// Serve static files (our UI) - this will be added at build time from our Nuxt frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticPath)))

	// CORS Middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}).Handler(r)

	p.logger.Debug("Listening on port 8077")
	http.ListenAndServe("0.0.0.0:8077", corsHandler)
}

func (p *program) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/config GET request")
	err := json.NewEncoder(w).Encode(p.config)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to encode config: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (p *program) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/config POST request")
	var newConfig Config
	json.NewDecoder(r.Body).Decode(&newConfig)
	// Write new config to file
	// This will later be saved in a db
}

func (p *program) handleListScripts(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/scripts GET request")
	files, _ := filepath.Glob(filepath.Join(p.scriptDir, "*.ps1"))
	json.NewEncoder(w).Encode(files)
}

func (p *program) handleAddScript(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/scripts POST request")
	// Logic to add new powershell scripts
}
