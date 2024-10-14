package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"

	"systray/icon"

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
	log.SetOutput(os.Stdout)
	log.Println("Starting WinSenseConnectSystray")
	loadConfig()
	go registerHotkeys()
	systray.Run(onReady, onExit)
}

func loadConfig() {
	log.Println("Loading configuration")
	file, err := os.ReadFile("systray_config.json")
	if err != nil {
		log.Printf("Error reading systray config file: %v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Error parsing systray config file: %v\n", err)
		os.Exit(1)
	}
	log.Printf("Loaded configuration: %+v\n", config)
}

func onReady() {
	log.Println("Systray is ready")
	log.Printf("Icon data length: %d bytes\n", len(icon.IconBytes))
	systray.SetIcon(icon.IconBytes)
	log.Println("Icon set")
	systray.SetTitle("WinSenseConnect Hotkeys")
	systray.SetTooltip("WinSenseConnect Hotkey Listener")

	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	go func() {
		<-mQuit.ClickedCh
		log.Println("Quit clicked, exiting")
		systray.Quit()
	}()
}

func onExit() {
	log.Println("Systray is exiting")
	hook.End()
}

func registerHotkeys() {
	log.Println("Registering hotkeys")
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
