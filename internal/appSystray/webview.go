package appSystray

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

func OpenURLInBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		log.Printf("Error opening URL in browser: %v", err)
		return fmt.Errorf("error opening URL in browser: %v", err)
	}

	return nil
}
