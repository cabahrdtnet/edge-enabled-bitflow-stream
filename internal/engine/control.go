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

	valueDescriptors = struct {
		IDs []string                // IDs of created Value Descriptors in EdgeX
		Initialized sync.WaitGroup  // synced access of receiving events after initializing value descriptors
	}{
		[]string{},
		sync.WaitGroup{}}
)

type configuration struct {
	EngineName   string
	Script       string
	Parameters   string
	InputTopic   string
	OutputTopic  string
	CommandTopic string
	ReverseCommandTopic string
	ReverseCommandResponseTopic string
	MqttBroker   string
}

// apply configuration for a run
func Configure() {
	valueDescriptors.Initialized.Add(1)
	go registerValueDescriptors()

	go initHeaderSubscription()
	go initEventSubscription()
	go initCommandSubscription()
	go initReverseCommandResponseSubscription()

	go handleCommand()
	go handlePublicationValue()
}

func CleanUp() {
	cleanUpValueDescriptors()
}

// initializes one-off subscription for first event
func initHeaderSubscription() {
	subscriber.event = subscribe(Config.InputTopic,
		Config.EngineName + "-event-subscriber",
		handleInitialEventMessage)
}

// initializes event subscription beginning with the second event
func initEventSubscription() {
	valueDescriptors.Initialized.Wait()
	subscriber.event = subscribe(Config.InputTopic,
		Config.EngineName + "-event-subscriber",
		handleEventMessage)
}

// initializes command subscription for commands from device service to device
func initCommandSubscription() {
	subscriber.command = subscribe(Config.CommandTopic,
		Config.EngineName + "-command-subscriber",
		handleCommandMessage)
}

// initializes reverse command subscription for commands from device to device service
func initReverseCommandResponseSubscription() {
	subscriber.reverseCommand = subscribe(Config.ReverseCommandResponseTopic,
		Config.EngineName + "-reverse-command-response-subscriber",
		handleReverseCommandMessage)
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

		msg := string(b)
		// TODO refactor and get from command line
		publish(Config.ReverseCommandTopic, Config.EngineName + "-reverse-command-publisher", msg)

		fmt.Println("Awaiting ID from server")
		response := <- reverseCommandResponse.incoming
		switch response {
		case "duplicate":
			fmt.Println("Value descriptor has already been registered.")
		case "error":
			panic("Value descriptor couldn't be created.")
		default:
			// response is ID of valueDescriptor
			fmt.Println("Adding value descriptor ID to value descriptors.")
			valueDescriptors.IDs = append(valueDescriptors.IDs, response)
		}
	}

	valueDescriptors.Initialized.Done()
}

func cleanUpValueDescriptors() {
	for _, ID := range valueDescriptors.IDs {
		reverseCommand := struct {
			Command string	`json:"command"`
			Payload string	`json:"payload"`
		}{
			"clean_value_descriptor",
			ID}

		b, err := json.Marshal(reverseCommand)
		msg := string(b)
		if err != nil {
			fmt.Println("Couldn't marshal reverse command message:", msg)
		}
		publish(Config.ReverseCommandTopic,
			Config.EngineName + "-reverse-command-publisher",
			msg)
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
			disconnect(subscriber.event)
			disconnect(subscriber.command)
			disconnect(subscriber.reverseCommand)
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
		msg := string(payload)
		publish(Config.OutputTopic, Config.EngineName + "-event-publisher", msg)
	}
	fmt.Println("events.outgoing is closed.")
}