package engine

import (
	"encoding/json"
	"fmt"
)

var (
	// these values are set in commands/device-bitflow/engine/main.go
	Config  = configuration{}
	// TODO rename: read comment below
	// both are event channels, where
	// publication == outgoing
	// subscription == incoming
	// data should be called events then
)

// TODO channels should contain only converted data, i.e. an event channel would be semantically better
// TODO implement json commands channel

// these values are set in commands/device-bitflow/engine/main.go
type configuration struct {
	Name         string
	Script       string
	Parameters   string
	InputTopic   string
	OutputTopic  string
	CommandTopic string
	MqttBroker   string
}

// apply configuration for a run
func Configure() {
	//go func() {
	//	for msg := range event.incoming {
	//		fmt.Println("Received: ", msg)
	//	}
	//}()

	go subscribeToData()
	go subscribeToCommand()

	go handleCommand()
	go handlePublicationValue()
}

func handleCommand() {
	for msg := range commands.incoming {
		// TODO this channel should contain custom JSON commands data
		switch msg {
		case "shutdown":
			fmt.Println("Closing channels...")
			close(events.incoming)
			close(commands.incoming)
			fmt.Println("Channels closed. Shutting down now.")
		default:
			fmt.Println("Ignoring unknown command.")
		}
	}
}

func handlePublicationValue() {
	for event := range events.outgoing {
		fmt.Println("PUBLISHING:", event)
		payload, err := json.Marshal(event)
		if err != nil {
			fmt.Println("Ignoring event: EdgeX Event can't be marshalled.")
		}
		message := string(payload)
		publish(message)
	}
	fmt.Println("events.outgoing is closed.")
}

// MQTT messages
// sub handler      writesTo  Message.incoming
// stdin of bitflow readsFrom Message.incoming

// processing
// stdout of bitflow writesTo  Message.publication
// handlePublicationValue           readsFrom Message.publication