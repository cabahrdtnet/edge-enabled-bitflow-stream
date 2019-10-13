package communication

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"os"
)

var (
	Broker = ""
)

// publish to topic `topic` as publisher `clientID` message `message`
func Publish(topic string, clientID string, msg string) {
	fmt.Printf("Publishing message `%s` as publisher `%s` to topic `%s`\n", msg, clientID, topic)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(Broker)
	opts.SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error().Error() + "[[Broker:" + Broker + "]]")
	}
	qos := 0
	token := client.Publish(topic, byte(qos), false, msg)
	token.Wait()

	client.Disconnect(250)
}

// subscribe to topic `topic` as subscriber `clientID` and call `handler` on received message
// boolean choke channel can be used to terminate the subscription
func Subscribe(topic string, clientID string, handler mqtt.MessageHandler) mqtt.Client {
	fmt.Printf("Subscribing to `%s` as `%s`.\n", topic, clientID)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(Broker)
	opts.SetClientID(clientID)

	opts.SetDefaultPublishHandler(handler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error().Error())
	}

	qos := 0
	if token := client.Subscribe(topic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	return client
}

// disconnect `client`
func Disconnect(client mqtt.Client) {
	client.Disconnect(250)
	optReader := client.OptionsReader()
	fmt.Printf("Cancel subscription of client `%s`.\n",
		optReader.ClientID())
}

