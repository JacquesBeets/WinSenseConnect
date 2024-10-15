package main

import (
	"fmt"
	"win-sense-connect/internal/bgService"

	"github.com/kardianos/service"
)

func main() {
	svcConfig := &service.Config{
		Name:        "WinSenseConnect",
		DisplayName: "MQTT Powershell Automation Service",
		Description: "Listens for MQTT messages and runs PowerShell scripts",
	}

	prg, err := bgService.NewProgram()
	if err != nil {
		fmt.Printf("Failed to create program: %v\n", err)
		return
	}
	defer prg.Logger.Close()

	s, err := service.New(prg, svcConfig)
	if err != nil {
		prg.Logger.Error(fmt.Sprintf("Failed to create service: %v", err))
		return
	}

	prg.Logger.Debug("Service created, running service...")

	err = s.Run()
	if err != nil {
		prg.Logger.Error(fmt.Sprintf("Service failed: %v", err))
		return
	}

	prg.Logger.Debug("Service run completed")
}
