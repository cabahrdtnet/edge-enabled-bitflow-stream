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

	subscriber = struct{
		event MQTT.Client
		command MQTT.Client
		reverseCommand MQTT.Client
	}{}
)

// publish to topic `topic` as publisher `clientID` message `message`
func publish(topic string, clientID string, msg string) {
	fmt.Printf("Publishing message `%s` as publisher `%s` to topic `%s`\n", msg, clientID, topic)
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(clientID)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	qos := 0
	token := client.Publish(topic, byte(qos), false, msg)
	token.Wait()

	client.Disconnect(250)
}

// subscribe to topic `topic` as subscriber `clientID` and call `handler` on received message
// boolean choke channel can be used to terminate the subscription
func subscribe(topic string, clientID string, handler MQTT.MessageHandler) MQTT.Client {
	fmt.Printf("Subscribing to `%s` as `%s`.\n", topic, clientID)
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(clientID)

	opts.SetDefaultPublishHandler(handler)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	qos := 0
	if token := client.Subscribe(topic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	return client
}

// disconnect `client`
func disconnect(client MQTT.Client) {
	client.Disconnect(250)
	optReader := client.OptionsReader()
	fmt.Printf("Cancel subscription of client `%s`.\n",
		optReader.ClientID())
}

// handles initial event messages that the input header is drawn from
func handleInitialEventMessage(client MQTT.Client, msg MQTT.Message) {
	// unmarshal from EdgeX JSON format to EdgeX Event
	event := models.Event{}
	payload := msg.Payload()
	err := json.Unmarshal(payload, &event)
	if err == nil {
		events.incoming <- event
	} else {
		fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
	}
	disconnect(subscriber.event)
}

// handle every event message beginning from the second event
func handleEventMessage(client MQTT.Client, msg MQTT.Message) {
	// unmarshal from EdgeX JSON format to EdgeX Event
	event := models.Event{}
	payload := msg.Payload()
	err := json.Unmarshal(payload, &event)
	if err == nil {
		events.incoming <- event
	} else {
		fmt.Println("Ignoring message: EdgeX Event JSON data can't be unmarshalled.")
	}
}

// handle command messages for commands from device service to device
func handleCommandMessage(client MQTT.Client, msg MQTT.Message) {
	commands.incoming <- string(msg.Payload())
}

// handle command messages for commands from service to device service
func handleReverseCommandMessage(client MQTT.Client, msg MQTT.Message) {
	reverseCommandResponse.incoming <- string(msg.Payload())
}