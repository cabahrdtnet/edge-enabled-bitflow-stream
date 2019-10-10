package engine

import (
	"encoding/json"
	"fmt"
	"github.com/datenente/device-bitflow/internal/communication"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	events = struct {
		outgoing chan models.Event
		incoming chan models.Event
	}{
		make(chan models.Event),
		make(chan models.Event)}

	initial = struct {
		outputHeader chan string
		processedEvent chan models.Event
	}{
		make(chan string),
		make(chan models.Event)}

	commands = struct {
		incoming chan string
	}{
		make(chan string)}

	// TODO move to outgoing commands of commands
	reverseCommandResponse = struct {
		incoming chan string
	}{
		make(chan string)}

	subscriber = struct{
		event MQTT.Client
		command MQTT.Client
		reverseCommand MQTT.Client
	}{}
)

// handles initial event messages that the input header is drawn from
func handleInitialEventMessage(client MQTT.Client, msg MQTT.Message) {
	// unmarshal from EdgeX JSON format to EdgeX Event
	event := models.Event{}
	payload := msg.Payload()
	err := json.Unmarshal(payload, &event)
	if err == nil {
		events.incoming <- event
	} else {
		fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
	}
	communication.Disconnect(subscriber.event)
}

// handle every event message beginning from the second event
func handleEventMessage(client MQTT.Client, msg MQTT.Message) {
	// unmarshal from EdgeX JSON format to EdgeX Event
	event := models.Event{}
	payload := msg.Payload()
	err := json.Unmarshal(payload, &event)
	if err == nil {
		events.incoming <- event
	} else {
		fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
	}
}

// handle command messages for commands from device service to device
func handleCommandMessage(client MQTT.Client, msg MQTT.Message) {
	commands.incoming <- string(msg.Payload())
}

// handle command messages for commands from service to device service
func handleReverseCommandMessage(client MQTT.Client, msg MQTT.Message) {
	reverseCommandResponse.incoming <- string(msg.Payload())
}