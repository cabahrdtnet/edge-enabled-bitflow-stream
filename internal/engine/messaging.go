package engine

import (
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"os"
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
)

// publish to topic `topic` as publisher `clientID` message `message`
func pub(topic string, clientID string, msg string) {
	fmt.Printf("Publishing message `%s` as publisher `%s` to topic `%s`", msg, clientID, topic)
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(clientID)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	qos := 0
	token := client.Publish(Config.OutputTopic, byte(qos), false, msg)
	token.Wait()

	client.Disconnect(250)
}

// subscribe to topic `topic` as subscriber `clientID` for `n` messages and call `handler` on received message
// negative numbers indicate infinite subscription
func subscribe(topic string, clientID string, n int, handler MQTT.MessageHandler) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.EngineName + "-event-subscriber")

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(handler)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	qos := 0
	if token := client.Subscribe(Config.InputTopic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	fmt.Println("Sample n-Subscriber Connected")

	// TODO replace with waitgroup and remove choke channel
	if n < 0 {
		for {
			<-choke
		}
	} else {
		for receiveCount < n {
			<-choke
			receiveCount++
		}
	}

	// TODO fix broker error msg: Socket error on client engine-0-event-subscriber, disconnecting
	client.Disconnect(250)
	fmt.Println("Sample n-Subscriber Disconnected")
}

// clientID: Config.EngineName + "-event-subscriber"
// topic:    Config.InputTopic
// n:        1
// handler:
//	func(client MQTT.Client, msg MQTT.Message) {
//		//choke <- [2]string{msg.Topic(), string(msg.Payload())}
//		// unmarshal from EdgeX JSON format to EdgeX Event
//		event := models.Event{}
//		payload := msg.Payload()
//		err := json.Unmarshal(payload, &event)
//		if err == nil {
//			events.incoming <- event
//		} else {
//			fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
//		}
//	}
func subscribeToDataOnce() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	// TODO rename this to something more meaningful
	opts.SetClientID(Config.EngineName + "-event-subscriber")
	//opts.SetUsername(*user)
	//opts.SetPassword(*password)

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
		// unmarshal from EdgeX JSON format to EdgeX Event
		event := models.Event{}
		payload := msg.Payload()
		err := json.Unmarshal(payload, &event)
		if err == nil {
			events.incoming <- event
		} else {
			fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
		}
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	qos := 0
	if token := client.Subscribe(Config.InputTopic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	fmt.Println("Sample 1-Subscriber Connected")

	num := 1
	for receiveCount < num {
		incoming := <-choke
		incoming = incoming
		fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming[0], incoming[1])
		receiveCount++
	}

	// TODO fix broker error msg: Socket error on client engine-0-event-subscriber, disconnecting
	client.Disconnect(250)
	fmt.Println("Sample 1-Subscriber Disconnected")
}

// clientID: Config.EngineName + "-event-subscriber"
// topic:    Config.InputTopic
// n:        -1
// handler:
//	func(client MQTT.Client, msg MQTT.Message) {
//		//choke <- [2]string{msg.Topic(), string(msg.Payload())}
//		// unmarshal from EdgeX JSON format to EdgeX Event
//		event := models.Event{}
//		payload := msg.Payload()
//		err := json.Unmarshal(payload, &event)
//		if err == nil {
//			events.incoming <- event
//		} else {
//			fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
//		}
//	}
func subscribeToData() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	// TODO rename this to something more meaningful
	opts.SetClientID(Config.EngineName + "-event-subscriber")
	//opts.SetUsername(*user)
	//opts.SetPassword(*password)

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		//choke <- [2]string{msg.Topic(), string(msg.Payload())}
		// unmarshal from EdgeX JSON format to EdgeX Event
		event := models.Event{}
		payload := msg.Payload()
		err := json.Unmarshal(payload, &event)
		if err == nil {
			events.incoming <- event
		} else {
			fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
		}
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	qos := 0
	if token := client.Subscribe(Config.InputTopic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	fmt.Println("Sample n-Subscriber Connected")

	num := 1000000
	for receiveCount < num {
		incoming := <-choke
		incoming = incoming
		//fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming[0], incoming[1])
		receiveCount++
	}

	// TODO fix broker error msg: Socket error on client engine-0-event-subscriber, disconnecting
	client.Disconnect(250)
	fmt.Println("Sample n-Subscriber Disconnected")
}

// clientID: Config.EngineName + "-event-publisher"
// topic:    Config.OutputTopic
func publish(payload string) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.EngineName + "-event-publisher")
	//opts.SetCleanSession(true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	num := 1
	qos := 0
	//payload := "This is a test and I'll see if it works."
	for i := 0; i < num; i++ {
		//fmt.Println("---- waiting for event to handlePublicationValue ----")
		//fmt.Println("---- doing handlePublicationValue ----", payload)
		token := client.Publish(Config.OutputTopic, byte(qos), false, payload)
		token.Wait()
	}

	client.Disconnect(250)
}

// clientID: Config.EngineName + "-reverse-command-publisher"
// topic:    "bitflow/engine/0/reverse-command"
func promptReverseCommand(payload string) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.EngineName + "-reverse-command-publisher")
	//opts.SetCleanSession(true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	num := 1
	qos := 0
	for i := 0; i < num; i++ {
		token := client.Publish("bitflow/engine/0/reverse-command", byte(qos), false, payload)
		token.Wait()
	}

	client.Disconnect(250)
}

// clientID: Config.EngineName + "-command-subscriber"
// topic:    Config.CommandTopic
// n:        -1
// handler:
//	func(client MQTT.Client, msg MQTT.Message) {
//		//choke <- [2]string{msg.Topic(), string(msg.Payload())}
//		// unmarshal from EdgeX JSON format to EdgeX Event
//		event := models.Event{}
//		payload := msg.Payload()
//		err := json.Unmarshal(payload, &event)
//		if err == nil {
//			events.incoming <- event
//		} else {
//			fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
//		}
//	}
func subscribeToCommand() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.EngineName + "-command-subscriber")
	//opts.SetUsername(*user)
	//opts.SetPassword(*password)

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
		//event.incoming <- string(msg.Payload()) //+"\n"
		commands.incoming <- string(msg.Payload())
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	qos := 0
	if token := client.Subscribe(Config.CommandTopic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	num := 1000000
	for receiveCount < num {
		incoming := <-choke
		//incoming = incoming
		fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming[0], incoming[1])
		receiveCount++
	}

	// TODO fix broker error msg: Socket error on client engine-0-commands-subscriber, disconnecting
	client.Disconnect(250)
	fmt.Println("Sample Subscriber Disconnected")
}


// clientID: Config.EngineName + "-reverse-command-response-subscriber"
// topic:    "bitflow/engine/0/reverse-command-response"
// n:        -1
// handler:
//	func(client MQTT.Client, msg MQTT.Message) {
//		//choke <- [2]string{msg.Topic(), string(msg.Payload())}
//		// unmarshal from EdgeX JSON format to EdgeX Event
//		event := models.Event{}
//		payload := msg.Payload()
//		err := json.Unmarshal(payload, &event)
//		if err == nil {
//			events.incoming <- event
//		} else {
//			fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
//		}
//	}
func subscribeToReverseCommandResponse() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.EngineName + "-reverse-command-response-subscriber")
	//opts.SetUsername(*user)
	//opts.SetPassword(*password)

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
		//event.incoming <- string(msg.Payload()) //+"\n"
		reverseCommandResponse.incoming <- string(msg.Payload())
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	qos := 0
	// TODO refactor this to use generic shit
	if token := client.Subscribe("bitflow/engine/0/reverse-command-response", byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	num := 1000000
	for receiveCount < num {
		incoming := <-choke
		//incoming = incoming
		fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming[0], incoming[1])
		receiveCount++
	}

	// TODO fix broker error msg: Socket error on client engine-0-commands-subscriber, disconnecting
	client.Disconnect(250)
	fmt.Println("Sample Subscriber Disconnected")
}