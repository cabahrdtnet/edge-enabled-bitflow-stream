package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type reverseCommandRequest struct {
	Command string                  `json:"command"`
	Payload models.ValueDescriptor  `json:"payload"`
}

type response struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func InitRegistrySubscription() {
	BitflowDriver.lc.Debug("Subscribing to " + naming.RegistryRequest)
	subscriber.registry = communication.Subscribe(
		naming.Topic(-1, naming.RegistryRequest),
		naming.Subscriber(-1, naming.RegistryRequest),
		handleRegistryRequestMessage)
}

// per device subscription for sink of engine
func InitSinkSubscription(index int64, events chan models.Event) {
	BitflowDriver.lc.Debug("Subscribing to " + naming.Sink + " of " + naming.Name(index))
	go handleSinkEvent(index, events)
	communication.Subscribe(
		naming.Topic(index, naming.Sink),
		naming.Subscriber(index, naming.Sink),
		func (client MQTT.Client, msg MQTT.Message) {
			event := models.Event{}
			payload := msg.Payload()
			err := json.Unmarshal(payload, &event)
			if err == nil {
				events <- event
			} else {
				BitflowDriver.lc.Debug("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
			}
		})
}

func handleSinkEvent(index int64, events chan models.Event) {
	for event := range events {
		url := config.URL.CoreData + clients.ApiEventRoute
		_, err := clients.PostJsonRequest(url, event, context.TODO())
		if err != nil {
			formatted := fmt.Sprintf("couldn't send event to core data: %v", err)
			BitflowDriver.lc.Debug(formatted)
		}
	}
	BitflowDriver.lc.Info("event channel of " + naming.Name(index) + " is closed")
}

// per device subscription for reverse command of engine
func InitReverseCommandSubscription(index int64, reverseCommands chan reverseCommandRequest) {
	BitflowDriver.lc.Debug("Subscribing to " + naming.ReverseCommand + " of " + naming.Name(index))
	go handleReverseCommand(index, reverseCommands)
	communication.Subscribe(
		naming.Topic(index, naming.ReverseCommand),
		naming.Subscriber(index, naming.ReverseCommand),
		func (client MQTT.Client, msg MQTT.Message) {
			reverseCommandRequest := reverseCommandRequest{}
			err := json.Unmarshal(msg.Payload(), &reverseCommandRequest)
			if err != nil {
				formatted := fmt.Sprintf("couldn't unmarshal reverse command request: %v", err)
				BitflowDriver.lc.Debug(formatted)
			}
			reverseCommands <- reverseCommandRequest
		})
}

func handleReverseCommand(index int64, reverseCommands chan reverseCommandRequest) {
	for reverseCommand := range reverseCommands {
		if reverseCommand.Command == "register_value_descriptor" {
			vd := reverseCommand.Payload
			url := config.URL.CoreData + clients.ApiValueDescriptorRoute
			ID, err := clients.PostJsonRequest(url, vd, context.TODO())
			if err != nil {
				formatted := fmt.Sprintf("couldn't register value descriptor in core data: %v", err)
				BitflowDriver.lc.Debug(formatted)
			}
			topic := naming.Topic(index, naming.ReverseCommandResponse)
			clientID := naming.Publisher(index, naming.ReverseCommandResponse)
			msg := ID
			communication.Publish(topic, clientID, msg)
			continue
		}

		if reverseCommand.Command == "clean_value_descriptor" {
			vd := reverseCommand.Payload
			name := vd.Name
			url := config.URL.CoreData + clients.ApiValueDescriptorRoute + "/name/" + name
			err := clients.DeleteRequest(url, context.TODO())
			if err != nil {
				formatted := fmt.Sprintf("couldn't clean value descriptor from core data: %v", err)
				BitflowDriver.lc.Debug(formatted)
			}
		}
	}
	BitflowDriver.lc.Info("reverse command channel of " + naming.Name(index) + " is closed")
}

func handleRegistryRequest() {
	for {
		select {
		case request := <- registration.request:
			var rsp response
			var err error
			var index int64

			if request > 0 {
				index = request
				err = register(index)
			} else {
				index = -request
				err = deregister(index)
			}

			if err == nil {
				rsp.Message = "success"
				rsp.Error = ""
			} else {
				rsp.Message = "failure"
				rsp.Error = err.Error()
			}

			payload, _ := json.Marshal(rsp)
			topic := naming.Topic(index, naming.RegistryResponse)
			clientID := naming.Publisher(index, naming.RegistryResponse)
			msg := string(payload)
			communication.Publish(topic, clientID, msg)

		case err := <- registration.err:
			var rsp response

			rsp.Message = "failure"
			rsp.Error = err.Error()

			payload, err := json.Marshal(rsp)
			if err != nil {
				BitflowDriver.lc.Error(err.Error())
			}
			msg := string(payload)
			BitflowDriver.lc.Error(msg)
			// TODO send this back to server
		}
	}
}