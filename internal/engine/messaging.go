package engine

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"
)

type dataChannel struct {
	publication  chan string
	subscription chan string
}

type commandChannel struct {
	Subscription chan string
}

func subscribeToData() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	// TODO rename this to something more meaningful
	opts.SetClientID(Config.Name + "-data-subscriber")
	//opts.SetUsername(*user)
	//opts.SetPassword(*password)

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		//choke <- [2]string{msg.Topic(), string(msg.Payload())}
		// unmarshal from EdgeX JSON format to EdgeX Event
		data.subscription <- string(msg.Payload()) //+"\n"
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

	num := 1000000
	for receiveCount < num {
		incoming := <-choke
		incoming = incoming
		//fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming[0], incoming[1])
		receiveCount++
	}

	// TODO fix broker error msg: Socket error on client engine-0-data-subscriber, disconnecting
	client.Disconnect(250)
	fmt.Println("Sample Subscriber Disconnected")
}

func publish(payload string) {
	fmt.Println("Sample Publisher Started")
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.Name + "-data-publisher")
	//opts.SetCleanSession(true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	num := 1
	qos := 0
	//payload := "This is a test and I'll see if it works."
	for i := 0; i < num; i++ {
		//fmt.Println("---- waiting for data to handlePublicationValue ----")
		//fmt.Println("---- doing handlePublicationValue ----", payload)
		token := client.Publish(Config.OutputTopic, byte(qos), false, payload)
		token.Wait()
	}

	client.Disconnect(250)
	fmt.Println("Sample Publisher Disconnected")
}

func subscribeToCommand() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.Name + "-command-subscriber")
	//opts.SetUsername(*user)
	//opts.SetPassword(*password)

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
		//data.subscription <- string(msg.Payload()) //+"\n"
		command.Subscription <- string(msg.Payload())
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

	// TODO fix broker error msg: Socket error on client engine-0-command-subscriber, disconnecting
	client.Disconnect(250)
	fmt.Println("Sample Subscriber Disconnected")
}