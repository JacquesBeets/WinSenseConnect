package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"win-sense-connect/internal/appSystray"
	"win-sense-connect/internal/appSystray/icon"
	"win-sense-connect/internal/common"
	"win-sense-connect/internal/shared"

	"github.com/getlantern/systray"
	hook "github.com/robotn/gohook"
)

var config common.SystrayConfig
var db *shared.DB

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

	log.Println("Initializing database connection")
	db, err = shared.NewDB()
	if err != nil {
		log.Printf("Error initializing database: %v\n", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v\n", err)
		}
	}()

	log.Println("Loading configuration")
	if err := loadConfig(); err != nil {
		log.Printf("Error loading configuration: %v\n", err)
		return
	}

	log.Println("Registering hotkeys")
	go registerHotkeys()

	log.Println("Starting systray")
	systray.Run(onReady, onExit)
}

func loadConfig() error {
	hotkeyCommands, err := db.GetHotkeyCommands()
	if err != nil {
		return fmt.Errorf("error getting hotkey commands from database: %v", err)
	}

	if len(hotkeyCommands) == 0 {
		log.Println("Warning: No hotkey commands found in the database")
	}

	config = common.SystrayConfig{
		HotkeyCommands: hotkeyCommands,
	}

	log.Printf("Loaded configuration with %d hotkey commands\n", len(config.HotkeyCommands))
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
				if err := appSystray.OpenURLInBrowser("http://localhost:8077"); err != nil {
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
	if err := cmd.Run(); err != nil {
		log.Printf("Error executing command: %v\n", err)
	}
}
