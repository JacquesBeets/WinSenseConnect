package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"win-sense-connect/internal/appSystray"
	"win-sense-connect/internal/appSystray/icon"

	"github.com/getlantern/systray"
	hook "github.com/robotn/gohook"
)

type HotkeyCommand struct {
	Hotkey  string `json:"hotkey"`
	Command string `json:"command"`
}

type SystrayConfig struct {
	HotkeyCommands []HotkeyCommand `json:"hotkeyCommands"`
}

var config SystrayConfig

func main() {
	// Set up logging to a file
	logFile, err := os.OpenFile("WinSenseConnectSystray.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("Starting WinSenseConnectSystray")

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in main: %v\n", r)
		}
	}()

	log.Println("Loading configuration")
	err = loadConfig()
	if err != nil {
		log.Printf("Error loading configuration: %v\n", err)
		return
	}

	log.Println("Registering hotkeys")
	go registerHotkeys()

	log.Println("Starting systray")
	systray.Run(onReady, onExit)
}

func loadConfig() error {
	file, err := os.ReadFile("systray_config.json")
	if err != nil {
		return fmt.Errorf("error reading systray config file: %v", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return fmt.Errorf("error parsing systray config file: %v", err)
	}
	log.Printf("Loaded configuration: %+v\n", config)
	return nil
}

func onReady() {
	log.Println("Systray is ready")
	log.Printf("Icon data length: %d bytes\n", len(icon.IconBytes))
	systray.SetIcon(icon.IconBytes)
	log.Println("Icon set")

	systray.SetTitle("WinSenseConnect Hotkeys")
	systray.SetTooltip("WinSenseConnect Hotkey Listener")

	mShowQuickShortcuts := systray.AddMenuItem("Dashboard", "Open Dashboard in Browser")
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	go func() {
		for {
			select {
			case <-mShowQuickShortcuts.ClickedCh:
				err := appSystray.OpenURLInBrowser("http://localhost:8077")
				if err != nil {
					log.Printf("Error opening shortcuts view: %v", err)
				}
			case <-mQuit.ClickedCh:
				log.Println("Quit clicked, exiting")
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	log.Println("Systray is exiting")
	hook.End()
}

func registerHotkeys() {
	for _, hc := range config.HotkeyCommands {
		keys := parseHotkey(hc.Hotkey)
		log.Printf("Registering hotkey: %s\n", hc.Hotkey)
		hook.Register(hook.KeyDown, keys, func(e hook.Event) {
			log.Printf("Hotkey pressed: %s\n", hc.Hotkey)
			go executeCommand(hc.Command)
		})
	}

	s := hook.Start()
	<-hook.Process(s)
}

func parseHotkey(hotkey string) []string {
	parts := strings.Split(hotkey, "+")
	var keys []string

	for _, part := range parts {
		switch strings.ToLower(part) {
		case "ctrl":
			keys = append(keys, "ctrl")
		case "alt":
			keys = append(keys, "alt")
		case "shift":
			keys = append(keys, "shift")
		case "win":
			keys = append(keys, "command")
		default:
			keys = append(keys, strings.ToLower(part))
		}
	}

	return keys
}

func executeCommand(command string) {
	log.Printf("Executing command: %s\n", command)
	cmd := exec.Command("cmd", "/C", command)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
	}
}
