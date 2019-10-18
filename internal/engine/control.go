package engine

import (
	"encoding/json"
	"fmt"
	"github.com/datenente/device-bitflow-stream/internal/communication"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"sync"
)

var (
	// these values are set in commands/device-bitflow-stream/engine/main.go
	Config  = configuration{}

	valueDescriptors = struct {
		IDs   []string       // ID of created Value Descriptors in EdgeX
		Names  []string       // Name of created Value Descriptors in EdgeX
		Initialized sync.WaitGroup // synced access of receiving events after initializing value descriptors
	}{
		[]string{},
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
	communication.Broker = Config.MqttBroker
	valueDescriptors.Initialized.Add(1)
	go registerValueDescriptors()

	go initHeaderSubscription()
	go initEventSubscription()
	go initCommandSubscription()
	go initReverseCommandResponseSubscription()

	go handleCommand()
	go handlePublicationValue()
}

// initializes one-off subscription for first event
func initHeaderSubscription() {
	subscriber.event = communication.Subscribe(Config.InputTopic,
		Config.EngineName + "-event-subscriber",
		handleInitialEventMessage)
}

// initializes event subscription beginning with the second event
func initEventSubscription() {
	valueDescriptors.Initialized.Wait()
	subscriber.event = communication.Subscribe(Config.InputTopic,
		Config.EngineName + "-event-subscriber",
		handleEventMessage)
}

// initializes command subscription for commands from device service to device
func initCommandSubscription() {
	subscriber.command = communication.Subscribe(Config.CommandTopic,
		Config.EngineName + "-command-subscriber",
		handleCommandMessage)
}

// initializes reverse command subscription for commands from device to device service
func initReverseCommandResponseSubscription() {
	subscriber.reverseCommand = communication.Subscribe(Config.ReverseCommandResponseTopic,
		Config.EngineName + "-reverse-command-response-subscriber",
		handleReverseCommandMessage)
}

// register value descriptors for EdgeX
func registerValueDescriptors() {
	initialProcessedEvent, ok := <- initial.processedEvent
	if !ok {
		panic("Initial processed event couldn't be retrieved, corresponding channel has been closed.")
	}
	vds := []contract.ValueDescriptor{}

	for _, reading := range initialProcessedEvent.Readings {
		vd := contract.ValueDescriptor{
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
			Command string                    `json:"command"`
			Payload contract.ValueDescriptor  `json:"payload"`
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
		communication.Publish(Config.ReverseCommandTopic, Config.EngineName + "-reverse-command-publisher", msg)

		// TODO remove support for erasing value descriptors
		// we can't know if there is not anyone else using the same descriptors
		// it is not possible as long as there are any readings in the event API
		// as we do not offer any support right now of moving these, we shouldn't remove value descriptors at all
		// if error, do not add to list
		// if it's a duplicate, we'll receive the ID of the value descriptor
		// if it's created, we'll receive the ID of the value descriptor

		// TODO stop cleaning this or so...
		fmt.Println("Awaiting ID from server")
		response := <- reverseCommandResponse.incoming
		switch response {
		case "error":
			panic("Value descriptor couldn't be created.")
		default:
			// response is ID of valueDescriptor
			fmt.Println("Adding value descriptor ID to value descriptors.")
			valueDescriptors.IDs = append(valueDescriptors.IDs, response)
			valueDescriptors.Names = append(valueDescriptors.Names, vd.Name)
		}
	}

	valueDescriptors.Initialized.Done()
}

func cleanUpValueDescriptors() {
	for _, name := range valueDescriptors.Names {
		reverseCommand := struct {
			Command string	`json:"command"`
			Payload contract.ValueDescriptor	`json:"payload"`
		}{
			"clean_value_descriptor",
			contract.ValueDescriptor{
				Name: name,
			},
		}

		b, err := json.Marshal(reverseCommand)
		msg := string(b)
		if err != nil {
			fmt.Println("Couldn't marshal clean_value_descriptor reverse command message:", msg)
		}
		communication.Publish(Config.ReverseCommandTopic,
			Config.EngineName + "-reverse-command-publisher",
			msg)
	}
	reverseCommand := struct {
		Command string	                 `json:"command"`
		Payload contract.ValueDescriptor `json:"payload"`
	}{
		"finalize_clean_value_descriptor",
		contract.ValueDescriptor{},
	}

	b, err := json.Marshal(reverseCommand)
	msg := string(b)
	if err != nil {
		fmt.Println("Couldn't marshal finalize reverse command message:", msg)
	}
	communication.Publish(Config.ReverseCommandTopic,
		Config.EngineName + "-reverse-command-publisher",
		msg)
}


func handleCommand() {
	for msg := range commands.incoming {
		fmt.Printf("%s::%d\n", msg, len(msg))
		switch msg {
		case "shutdown":
			fmt.Println("Closing channels...")
			close(events.incoming)
			close(commands.incoming)
			communication.Disconnect(subscriber.event)
			communication.Disconnect(subscriber.command)
			communication.Disconnect(subscriber.reverseCommand)

		//case "deregister":
		//	fmt.Println("Closing channels...")
		//	close(events.incoming)
		//	close(commands.incoming)
		//	communication.Disconnect(subscriber.event)
		//	communication.Disconnect(subscriber.command)
		//	cleanUpValueDescriptors()
		//	communication.Disconnect(subscriber.reverseCommand)
		//	fmt.Println("Channels closed and deregistered value descriptors. Shutting down now.")

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
		communication.Publish(Config.OutputTopic, Config.EngineName + "-event-publisher", msg)
	}
	fmt.Println("events.outgoing is closed.")
}