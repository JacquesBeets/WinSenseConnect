package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

type HotkeyCommand struct {
	Hotkey  string `json:"hotkey"`
	Command string `json:"command"`
}

type SystrayConfig struct {
	HotkeyCommands []HotkeyCommand `json:"hotkeyCommands"`
}

var (
	user32                                      = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey                          = user32.NewProc("RegisterHotKey")
	procGetMessage                              = user32.NewProc("GetMessageW")
	modControl, modAlt, modShift, modWin uint32 = 0x0002, 0x0001, 0x0004, 0x0008
)

const (
	wmHotkey = 786
)

type MSG struct {
	HWND   uintptr
	UINT   uint32
	WPARAM uintptr
	LPARAM uintptr
	DWORD  uint32
	POINT  struct{ X, Y int32 }
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
	file, err := ioutil.ReadFile("systray_config.json")
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
	systray.SetIcon(getIcon())
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
}

func getIcon() []byte {
	// Replace this with actual icon data
	return []byte{0}
}

func registerHotkeys() {
	log.Println("Registering hotkeys")
	for i, hc := range config.HotkeyCommands {
		modifiers, key := parseHotkey(hc.Hotkey)
		r, _, err := procRegisterHotKey.Call(0, uintptr(i), uintptr(modifiers), uintptr(key))
		if r == 0 {
			log.Printf("Failed to register hotkey: %s, error: %v\n", hc.Hotkey, err)
		} else {
			log.Printf("Registered hotkey: %s\n", hc.Hotkey)
		}
	}

	var msg MSG
	for {
		r, _, err := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if r == 0 {
			log.Printf("GetMessage failed: %v\n", err)
			return
		}

		if msg.UINT == wmHotkey {
			id := int(msg.WPARAM)
			if id >= 0 && id < len(config.HotkeyCommands) {
				log.Printf("Hotkey pressed: %s\n", config.HotkeyCommands[id].Hotkey)
				go executeCommand(config.HotkeyCommands[id].Command)
			}
		}
	}
}

func parseHotkey(hotkey string) (uint32, uint32) {
	parts := strings.Split(hotkey, "+")
	var modifiers uint32
	var key uint32

	for _, part := range parts {
		switch strings.ToLower(part) {
		case "ctrl":
			modifiers |= modControl
		case "alt":
			modifiers |= modAlt
		case "shift":
			modifiers |= modShift
		case "win":
			modifiers |= modWin
		default:
			key = uint32(part[0])
		}
	}

	return modifiers, key
}

func executeCommand(command string) {
	log.Printf("Executing command: %s\n", command)
	cmd := exec.Command("cmd", "/C", command)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
	}
}

func SetProcessDPIAware() {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("SetProcessDPIAware")
	proc.Call()
}

func init() {
	SetProcessDPIAware()
}
