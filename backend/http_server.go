package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
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
	r.HandleFunc("/api/scripts/{id}", p.handleGetScript).Methods("GET")
	r.HandleFunc("/api/scripts", p.handleAddScript).Methods("POST")
	r.HandleFunc("/api/restart", p.handleRestartService).Methods("POST")
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
	err := json.NewDecoder(r.Body).Decode(&newConfig)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to decode config: %v", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	p.config = newConfig
	err = p.db.UpdateConfig(&p.config)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to save config: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (p *program) handleListScripts(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/scripts GET request")

	scritpConfig, err := p.db.GetScriptConfigs()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to get script configs: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(scritpConfig)
}

func (p *program) handleGetScript(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/scripts/:id GET request")
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to parse id: %v", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	scriptConfig, err := p.db.GetScriptConfig(id)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to get script config: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(scriptConfig)
}

func (p *program) handleAddScript(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/scripts POST request")
	// Logic to add new powershell scripts
}

func (p *program) handleRestartService(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug("Handling /api/restart POST request")
	err := p.restartService()
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to restart service: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
