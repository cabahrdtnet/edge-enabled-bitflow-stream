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
	// TODO IDs from server, which should later be removed from it
	valueDescriptorsIDs         = []string{}
	valueDescriptorsInitialized sync.WaitGroup
	bitflowPipeline             sync.WaitGroup
)

// TODO channels should contain only converted data, i.e. an event channel would be semantically better
// TODO implement json commands channel

// these values are set in commands/device-bitflow/engine/main.go
type configuration struct {
	EngineName   string
	Script       string
	Parameters   string
	InputTopic   string
	OutputTopic  string
	CommandTopic string
	MqttBroker   string
}

// apply configuration for a run
func Configure() {
	setup()
	go registerValueDescriptors()

	go subscribeToCommand()
	go subscribeToReverseCommandResponse()

	go handleCommand()
	go handlePublicationValue()
}

func setup() {
	valueDescriptorsInitialized.Add(1)
	bitflowPipeline.Add(1)
	go subscribeToDataOnce()
}

func CleanUp() {
	cleanUpValueDescriptors()
}

// register value descriptors for EdgeX
func registerValueDescriptors() {
	initialProcessedEvent, ok := <- initial.processedEvent
	if !ok {
		panic("Initial processed event couldn't be retrieved, corresponding channel has been closed.")
	}
	vds := []models.ValueDescriptor{}

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
				"created-by-" + Config.EngineName,
			},
		}
		vds = append(vds, vd)
	}

	// TODO send value descriptor to server one by one, as you need to get an answer for each
	for _, vd := range vds {
		reverseCommand := struct {
			Command string                  `json:"command"`
			Payload models.ValueDescriptor  `json:"payload"`
		}{
			"register_value_descriptor",
			vd}

		b, err := json.Marshal(reverseCommand)
		if err != nil {
			fmt.Println("Couldn't marshal reverse command packet:", string(b))
		}

		fmt.Println("Prompt register_value_descriptor reverse command.")
		promptReverseCommand(string(b))
		fmt.Println("Awaiting ID from server")
		response := <- reverseCommandResponse.incoming
		// TODO duplicates are fine; ID needs to be saved for later removal; other responses should result in error
		switch response {
		case "duplicate":
			fmt.Println("Value descriptor has already been registered.")
		case "error":
			panic("Value descriptor couldn't be created.")
		default:
			// response is ID of valueDescriptor
			fmt.Println("Adding value descriptor ID to value descriptors.")
			valueDescriptorsIDs = append(valueDescriptorsIDs, response)
		}
	}

	valueDescriptorsInitialized.Done()
	subscribeToData()
}

func cleanUpValueDescriptors() {
	for _, ID := range valueDescriptorsIDs {
		reverseCommand := struct {
			Command string	`json:"command"`
			Payload string	`json:"payload"`
		}{
			"clean_value_descriptors",
			ID}

		b, err := json.Marshal(reverseCommand)
		if err != nil {
			fmt.Println("Couldn't marshal reverse command packet:", string(b))
		}
		promptReverseCommand(string(b))
	}
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