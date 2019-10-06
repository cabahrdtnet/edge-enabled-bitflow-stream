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

func publish(payload string) {
	fmt.Println("Sample Publisher Started")
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
	fmt.Println("Sample Publisher Disconnected")
}

func promptReverseCommand(payload string) {
	fmt.Println("Sample Publisher Started")
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
	fmt.Println("Sample Publisher Disconnected")
}

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