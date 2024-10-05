package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

type SensorData struct {
	Timestamp      time.Time                 `json:"timestamp"`
	CPUUsage       float64                   `json:"cpu_usage"`
	CPUInfo        cpu.InfoStat              `json:"cpu_info"`
	MemoryUsage    float64                   `json:"memory_usage"`
	DiskUsage      float64                   `json:"disk_usage"`
	Uptime         uint64                    `json:"uptime"`
	DiskPartitions []disk.PartitionStat      `json:"disk_partitions"`
	NetConnections []net.ConnectionStat      `json:"net_connections"`
	Users          []host.UserStat           `json:"users"`
	Sensors        []sensors.TemperatureStat `json:"sensors"`
	CPUTemperature float64                   `json:"cpu_temperature"`
}

type TemperatureData struct {
	Temperature float64 `json:"Temperature"`
}

func collectSensorData() (SensorData, error) {
	data := SensorData{
		Timestamp: time.Now(),
	}

	// CPU Usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		data.CPUUsage = cpuPercent[0]
	}

	// CPU Info
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		data.CPUInfo = cpuInfo[0]
	}

	// Memory Usage
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		data.MemoryUsage = memInfo.UsedPercent
	}

	// Disk Usage
	diskInfo, err := disk.Usage("C:")
	if err == nil {
		data.DiskUsage = diskInfo.UsedPercent
	}

	// Disk Partitions
	partitions, err := disk.Partitions(false)
	if err == nil {
		data.DiskPartitions = partitions
	}

	// Net
	netConnections, err := net.Connections("all")
	if err == nil {
		data.NetConnections = netConnections
	}

	// Uptime
	hostInfo, err := host.Info()
	if err == nil {
		data.Uptime = hostInfo.Uptime
	}

	// Users
	users, err := host.Users()
	if err == nil {
		data.Users = users
	}

	// Sensors
	sensors, err := sensors.TemperaturesWithContext(context.Background())
	if err == nil {
		data.Sensors = sensors
	}

	// CPU Temperature
	temp, err := getCPUTemperature()
	if err == nil {
		data.CPUTemperature = temp
	} else {
		fmt.Printf("Failed to get CPU temperature: %v\n", err)
	}

	return data, nil
}

func getCPUTemperature() (float64, error) {
	cmd := exec.Command("powershell", "-Command", `
		$temp = Get-WmiObject MSAcpi_ThermalZoneTemperature -Namespace "root/wmi" | Select-Object -First 1
		if ($temp) {
			$celsius = ($temp.CurrentTemperature / 10) - 273.15
			ConvertTo-Json @{ Temperature = [math]::Round($celsius, 2) }
		} else {
			ConvertTo-Json @{ Temperature = $null }
		}
	`)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute PowerShell command: %v", err)
	}

	var tempData TemperatureData
	err = json.Unmarshal(output, &tempData)
	if err != nil {
		return 0, fmt.Errorf("failed to parse temperature data: %v", err)
	}

	if tempData.Temperature == 0 {
		return 0, fmt.Errorf("temperature data not available")
	}

	return tempData.Temperature, nil
}
