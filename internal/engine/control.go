package engine

import "fmt"

var (
	// these values are set in command/device-bitflow/engine/main.go
	Config  = configuration{}
	// TODO rename: read comment below
	// both are event channels, where
	// publication == outgoing
	// subscription == incoming
	// data should be called events then
	data    = dataChannel{make(chan string), make(chan string)}
	command = commandChannel{make(chan string)}
)

// TODO channels should contain only converted data, i.e. an event channel would be semantically better
// TODO implement json command channel

// these values are set in command/device-bitflow/engine/main.go
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
	//	for msg := range data.subscription {
	//		fmt.Println("Received: ", msg)
	//	}
	//}()

	go subscribeToData()
	go subscribeToCommand()

	go handleCommand()
	go handlePublicationValue()
}

func handleCommand() {
	for msg := range command.Subscription {
		// TODO this channel should contain custom JSON command data
		switch msg {
		case "shutdown":
			fmt.Println("Closing channels...")
			close(data.subscription)
			close(command.Subscription)
			fmt.Println("Channels closed. Shutting down now.")
		default:
			fmt.Println("Ignoring unknown command.")
		}
	}
}

// TODO who's doing the conversion and whose marshalling? it's also a perspective question
func handlePublicationValue() {
	for payload := range data.publication {
		// TODO marshal here?
		// marshal from EdgeX event to EdgeX JSON format
		fmt.Println("PUBLISHING:", payload)
		publish(payload)
	}
	fmt.Println("data.publication is closed.")
}

// MQTT messages
// sub handler      writesTo  Message.subscription
// stdin of bitflow readsFrom Message.subscription

// processing
// stdout of bitflow writesTo  Message.publication
// handlePublicationValue           readsFrom Message.publication