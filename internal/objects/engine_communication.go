package objects

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type EngineCommunication struct {
	Index                    int64       // communication should belong to engine with the same index
	SinkSubscriber           mqtt.Client // mqtt client that connects from events of engine, e.g. "bitflow/engine/1/sink"
	ReverseCommandSubscriber mqtt.Client // mqtt client that connects to reverse commands from engine, e.g. "bitflow/engine/1/reverse-command"
	Events                   chan contract.Event
	ReverseCommandRequests   chan reverseCommandRequest
	ValueDescriptorsCleaned  chan bool
}

type reverseCommandRequest struct {
	Command string                   `json:"command"`
	Payload contract.ValueDescriptor `json:"payload"`
}

// open channels and init subscriptions for engine with this EngineCommunication
func (ec *EngineCommunication) Setup() {
	ec.Events = make(chan contract.Event)
	ec.ReverseCommandRequests = make(chan reverseCommandRequest)
	ec.ValueDescriptorsCleaned = make(chan bool)
	go ec.initEventSubscription()
	go ec.initReverseCommandSubscription()
	go ec.handleEvent()
	go ec.handleReverseCommand()
}

// close channels and cancel subscriptions for engine with this EngineCommunication
func (ec *EngineCommunication) Teardown() {
	close(ec.Events)
	close(ec.ReverseCommandRequests)
	communication.Disconnect(ec.SinkSubscriber)
	communication.Disconnect(ec.ReverseCommandSubscriber)
	close(ec.ValueDescriptorsCleaned)
}

// init event subscription
func (ec *EngineCommunication) initEventSubscription() {
	config.Log.Debug("Subscribing to " + naming.Sink + " of " + naming.Name(ec.Index))
	ec.SinkSubscriber = communication.Subscribe(
		naming.Topic(ec.Index, naming.Sink),
		naming.Subscriber(ec.Index, naming.Sink),
		func (client mqtt.Client, msg mqtt.Message) {
			event := contract.Event{}
			payload := msg.Payload()
			err := json.Unmarshal(payload, &event)
			if err == nil {
				ec.Events <- event
			} else {
				config.Log.Debug("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
			}
		})
}

// handle event from engine
func (ec *EngineCommunication) handleEvent() {
	for event := range ec.Events {
		url := config.URL.CoreData + clients.ApiEventRoute
		_, err := clients.PostJsonRequest(url, event, context.TODO())
		if err != nil {
			formatted := fmt.Sprintf("couldn't send event to core data: %v", err)
			config.Log.Debug(formatted)
		}
	}
	config.Log.Info("event channel of " + naming.Name(ec.Index) + " is closed")
}

// per device subscription for reverse command of engine
func (ec *EngineCommunication) initReverseCommandSubscription() {
	config.Log.Debug("Subscribing to " + naming.ReverseCommand + " of " + naming.Name(ec.Index))
	ec.ReverseCommandSubscriber = communication.Subscribe(
		naming.Topic(ec.Index, naming.ReverseCommand),
		naming.Subscriber(ec.Index, naming.ReverseCommand),
		func (client mqtt.Client, msg mqtt.Message) {
			reverseCommandRequest := reverseCommandRequest{}
			err := json.Unmarshal(msg.Payload(), &reverseCommandRequest)
			if err != nil {
				formatted := fmt.Sprintf("couldn't unmarshal reverse command request: %v", err)
				config.Log.Debug(formatted)
			}
			ec.ReverseCommandRequests <- reverseCommandRequest
		})
}

// handle reverse command from engine
func (ec *EngineCommunication) handleReverseCommand() {
	for rcr := range ec.ReverseCommandRequests {
		if rcr.Command == "register_value_descriptor" {
			vd := rcr.Payload
			url := config.URL.CoreData + clients.ApiValueDescriptorRoute
			ID, err := clients.PostJsonRequest(url, vd, context.TODO())
			msg := ID
			if err != nil && err.(*types.ErrServiceClient).StatusCode == 409 {
				formatted := fmt.Sprintf("couldn't register value descriptor: %v", err)
				config.Log.Debug(formatted)
				url += "/name/" + vd.Name
				payload, err := clients.GetRequest(url, context.TODO())
				if err != nil {
					formatted := fmt.Sprintf("couldn't get already registered value descriptor: %v", err)
					config.Log.Debug(formatted)
					msg = "error"
				}
				var alreadyRegisteredValueDescriptor contract.ValueDescriptor
				err = json.Unmarshal(payload, &alreadyRegisteredValueDescriptor)
				if err != nil {
					formatted := fmt.Sprintf("couldn't marshal already registered value descriptor: %v", err)
					config.Log.Debug(formatted)
					msg = "error"
				}
				msg = alreadyRegisteredValueDescriptor.Id
			}

			if err != nil && err.(*types.ErrServiceClient).StatusCode != 409 {
				formatted := fmt.Sprintf("couldn't register value descriptor: %v", err)
				config.Log.Debug(formatted)
				msg = "error"
			}

			topic := naming.Topic(ec.Index, naming.ReverseCommandResponse)
			clientID := naming.Publisher(ec.Index, naming.ReverseCommandResponse)
			communication.Publish(topic, clientID, msg)
			continue
		}

		if rcr.Command == "clean_value_descriptor" {
			vd := rcr.Payload
			name := vd.Name
			url := config.URL.CoreData + clients.ApiValueDescriptorRoute + "/name/" + name
			err := clients.DeleteRequest(url, context.TODO())
			if err != nil {
				formatted := fmt.Sprintf("couldn't clean value descriptor from core data: %v", err)
				config.Log.Debug(formatted)
			}
			continue
		}

		if rcr.Command == "finalize_clean_value_descriptor" {
			ec.ValueDescriptorsCleaned <- true
		}
	}
	config.Log.Info("reverse command channel of " + naming.Name(ec.Index) + " is closed")
}

