package engine

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"sync"
)

var (
	// these values are set in commands/device-bitflow/engine/main.go
	Config  = configuration{}
	// TODO rename: read comment below
	// both are event channels, where
	// publication == outgoing
	// subscription == incoming
	// data should be called events then
	valueDescriptorsInitialized sync.WaitGroup
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
	valueDescriptorsInitialized.Add(1)

	go subscribeToDataOnce()
	go registerValueDescriptors()
	go subscribeToCommand()
	go subscribeToReverseCommandResponse()

	go handleCommand()
	go handlePublicationValue()
}

// value descriptors for EdgeX
func registerValueDescriptors() {
	initialProcessedEvent, ok := <- initial.processedEvent
	if !ok {
		panic("Initial processed event couldn't be retrieved, corresponding channel has been closed.")
	}
	vds := []models.ValueDescriptor{}

	// - derive value descriptors via created readings and output header
	// (time,tags,)humancount,humancount_avg,caninecount,caninecount_avg
	for _, reading := range initialProcessedEvent.Readings {
		vd := models.ValueDescriptor{
			Name:
				reading.Name,
			Description:
				fmt.Sprintf("auto generated value by bitflow script execution engine named: %s",
					reading.Name),
			Type:
				typeOf(reading.Value),
			UomLabel:      "",
			Formatting:    "%s",
			Labels:        []string{
				"bitflow-value-descriptor",
				"created-by-" + reading.Device,
			},
		}
		vds = append(vds, vd)
	}

	// TODO send value descriptor to server one by one, as you need to get an answer for each
	// - marshal created VD and request DS
	payload, err := json.Marshal(vds)
	if err != nil {
		fmt.Println("Couldn't marshal value descriptor slice.")
	}
	// - publish message over ReverseCommand topic
	// - let them send to metadata
	promptReverseCommand(string(payload))

	// - await response
	response := <- reverseCommandResponse.incoming
	if response != "ok" {
		panic(response)
	}
	valueDescriptorsInitialized.Done()
	subscribeToData()
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