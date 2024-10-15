package bgService

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
	p.Logger.Debug("Starting HTTP server")
	r := p.router

	exePath, err := os.Executable()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to get executable path: %v", err))
	}
	staticPath := filepath.Join(filepath.Dir(exePath), "frontend/.output/public")

	// API endpoints
	r.HandleFunc("/api/config", p.handleGetConfig).Methods("GET")
	r.HandleFunc("/api/config", p.handleUpdateConfig).Methods("POST")
	r.HandleFunc("/api/scripts", p.handleListScripts).Methods("GET")
	r.HandleFunc("/api/scripts/{id}", p.handleGetScript).Methods("GET")
	r.HandleFunc("/api/scripts", p.handleAddScript).Methods("POST")
	r.HandleFunc("/api/restart", p.handleRestartService).Methods("POST")
	r.HandleFunc("/api/events", p.eventHandler)
	r.HandleFunc("/api/logs", p.handleGetLogs).Methods("GET")
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

	p.Logger.Debug("Listening on port 8077")
	http.ListenAndServe("0.0.0.0:8077", corsHandler)
}

func (p *program) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/config GET request")
	err := json.NewEncoder(w).Encode(p.config)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to encode config: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (p *program) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/config POST request")
	var newConfig Config
	err := json.NewDecoder(r.Body).Decode(&newConfig)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to decode config: %v", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	p.config = newConfig
	err = p.db.UpdateConfig(&p.config)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to save config: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (p *program) handleListScripts(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/scripts GET request")

	scritpConfig, err := p.db.GetScriptConfigs()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to get script configs: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(scritpConfig)
}

func (p *program) handleGetScript(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/scripts/:id GET request")
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to parse id: %v", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	scriptConfig, err := p.db.GetScriptConfig(id)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to get script config: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(scriptConfig)
}

func (p *program) handleAddScript(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/scripts POST request")
	// Logic to add new powershell scripts
}

func (p *program) handleRestartService(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/restart POST request")
	err := p.restartService()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to restart service: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (p *program) eventHandler(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/events SSE Events")
	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new channel for this client
	events := make(chan []byte)
	p.eventChannels = append(p.eventChannels, events)
	defer func() {
		// Remove this client's channel when the connection is closed
		for i, ch := range p.eventChannels {
			if ch == events {
				p.eventChannels = append(p.eventChannels[:i], p.eventChannels[i+1:]...)
				break
			}
		}
		close(events)
	}()

	// Stream events to the client
	for {
		select {
		case event := <-events:
			fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (p *program) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	p.Logger.Debug("Handling /api/logs GET request")
	exePath, err := os.Executable()
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to get executable path: %v", err))
	}
	dir := filepath.Join(filepath.Dir(exePath), "WinSenseConnect.log")
	// Read the log file
	logContent, err := os.ReadFile(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, return an empty log array
			p.Logger.Debug("Log file not found. Returning empty log array.")
			json.NewEncoder(w).Encode(map[string]string{"logs": ""})
			return
		}
		p.Logger.Error(fmt.Sprintf("Failed to read log file: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create a response struct
	response := struct {
		Logs string `json:"logs"`
	}{
		Logs: string(logContent),
	}

	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to encode log response: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
