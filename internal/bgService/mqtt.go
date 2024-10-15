package bgService

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	topicBase           = "winsense/"
	configTopic         = ""
	configResponseTopic = ""
)

func (p *program) onConnect(client mqtt.Client) {
	defer func() {
		if r := recover(); r != nil {
			p.Logger.Error(fmt.Sprintf("Recovered from panic in onConnect: %v\nStack trace: %s", r, debug.Stack()))
		}
	}()

	p.Logger.Debug("Connected to MQTT broker")

	// Set topics
	configTopic = topicBase + p.config.Topic + "/" + p.config.ClientID
	configResponseTopic = topicBase + p.config.Topic + "/" + p.config.ClientID + "/response"

	// Subscribe to the command topic
	if token := client.Subscribe(configTopic, 0, p.commandHandler); token.Wait() && token.Error() != nil {
		errMsg := fmt.Sprintf("Failed to subscribe to command topic: %v", token.Error())
		p.Logger.Error(errMsg)
	} else {
		p.Logger.Debug(fmt.Sprintf("Successfully subscribed to command topic: %s", configTopic))
	}

	// Subscribe to the response topic
	if token := client.Subscribe(configResponseTopic, 0, p.responseHandler); token.Wait() && token.Error() != nil {
		errMsg := fmt.Sprintf("Failed to subscribe to response topic: %v", token.Error())
		p.Logger.Error(errMsg)
	} else {
		p.Logger.Debug(fmt.Sprintf("Successfully subscribed to response topic: %s", configResponseTopic))
	}
}

func (p *program) onConnectionLost(client mqtt.Client, err error) {
	p.Logger.Error(fmt.Sprintf("Connection to MQTT broker lost: %v", err))
}

func (p *program) commandHandler(client mqtt.Client, msg mqtt.Message) {
	defer func() {
		if r := recover(); r != nil {
			p.Logger.Error(fmt.Sprintf("Recovered from panic in commandHandler: %v\nStack trace: %s", r, debug.Stack()))
		}
	}()

	command := string(msg.Payload())
	p.Logger.Debug(fmt.Sprintf("Received command: %s", command))

	scriptConfig, exists := p.config.Commands[command]
	if !exists {
		p.Logger.Error(fmt.Sprintf("Unknown command: %s", command))
		return
	}

	scriptPath := filepath.Join(p.scriptDir, scriptConfig.ScriptPath)
	p.Logger.Debug(fmt.Sprintf("Executing script: %s", scriptPath))

	output, err := p.executeScript(scriptPath, scriptConfig.RunAsUser)
	if err != nil {
		errMsg := fmt.Sprintf("Error executing script for command '%s': %v", command, err)
		p.Logger.Error(errMsg)
		p.publishResponse(client, errMsg)
	} else {
		p.Logger.Debug(fmt.Sprintf("Successfully executed command: %s\nOutput: %s", command, output))
		p.publishResponse(client, output)
	}
}

func (p *program) responseHandler(client mqtt.Client, msg mqtt.Message) {
	p.Logger.Debug(fmt.Sprintf("Received response: %s", string(msg.Payload())))
}

func (p *program) publishResponse(client mqtt.Client, message string) {
	if token := client.Publish(configResponseTopic, 0, false, message); token.Wait() && token.Error() != nil {
		p.Logger.Error(fmt.Sprintf("Failed to publish script output: %v", token.Error()))
	}
}

func (p *program) publishSensorData() {
	// TODO: Uncomment this when sensors are implemented
	// Must be able to manage multiple sensors independently

	// ticker := time.NewTicker(time.Duration(p.config.SensorConfig.Interval) * time.Second)
	// defer ticker.Stop()

	// for range ticker.C {
	// 	sensorData, err := collectSensorData()
	// 	if err != nil {
	// 		p.Logger.Error(fmt.Sprintf("Failed to collect sensor data: %v", err))
	// 		continue
	// 	}

	// 	jsonData, err := json.Marshal(sensorData)
	// 	if err != nil {
	// 		p.Logger.Error(fmt.Sprintf("Failed to marshal sensor data: %v", err))
	// 		continue
	// 	}

	// 	token := p.mqttClient.Publish(p.config.SensorConfig.SensorTopic, 0, false, jsonData)
	// 	if token.Wait() && token.Error() != nil {
	// 		p.Logger.Error(fmt.Sprintf("Failed to publish sensor data: %v", token.Error()))
	// 	} else {
	// 		p.Logger.Debug("Successfully published sensor data")
	// 	}
	// }
}

func (p *program) setupMQTTClient() {
	opts := mqtt.NewClientOptions().AddBroker(p.config.BrokerAddress)
	opts.SetClientID(p.config.ClientID)
	opts.SetUsername(p.config.Username)
	opts.SetPassword(p.config.Password)
	opts.SetOnConnectHandler(p.onConnect)
	opts.SetConnectionLostHandler(p.onConnectionLost)

	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Minute * 5)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(time.Second * 10)

	p.mqttClient = mqtt.NewClient(opts)

	if p.config.SensorConfigEnabled {
		go p.publishSensorData()
	}
}
