package engine

import "fmt"

var (
	// these values are set in cmd/device-bitflow/engine/main.go
	Config = configuration{}
	data   = dataChannel{make(chan string), make(chan string)}
	cmd    = commandChannel{make(chan string)}
	closing = false
)

// these values are set in cmd/device-bitflow/engine/main.go
type configuration struct {
	Name         string
	Script       string
	Arguments    string
	InputTopic   string
	OutputTopic  string
	CommandTopic string
	MqttBroker   string
}

// apply configuration for a run
func Configure() {
	//go func() {
	//	for msg := range data.Subscription {
	//		fmt.Println("Received: ", msg)
	//	}
	//}()

	go InitSubscription()
	go subscribeCommand()

	go handleCommand()
	go handlePublicationValue()
}

// TODO move, handlePublicationValue, subscribe here
func handleCommand() {
	for msg := range cmd.Subscription {
		switch msg {
		case "shutdown":
			fmt.Println("Closing channels...")
			close(data.Subscription)
			close(cmd.Subscription)
			//close(data.Publication)
			fmt.Println("Channels closed. Shutting down now.")
		default:
			fmt.Println("Ignoring unknown command.")
		}
	}
}

func handlePublicationValue() {
	for payload := range data.Publication {
		//fmt.Println("hello")
		// foo, ok := <- ch
		//             if !ok {
		//                println("done")
		//                wg.Done()
		//                return
		//            }
		// TODO handle: empty data
		fmt.Println("PUBLISHING:", payload)
		publish(payload)
		//fmt.Println("byebye")
	}
	fmt.Println("data.publication is closed.")
}

// MQTT messages
// sub handler      writesTo  Message.Subscription
// stdin of bitflow readsFrom Message.Subscription

// processing
// stdout of bitflow writesTo  Message.Publication
// handlePublicationValue           readsFrom Message.Publication