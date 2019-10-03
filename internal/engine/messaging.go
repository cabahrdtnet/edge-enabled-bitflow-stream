package engine

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"
)

type messenger struct {
	Publication chan string
	Subscription chan string
}

func InitSubscription() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.Name)
	//opts.SetUsername(*user)
	//opts.SetPassword(*password)

	receiveCount := 0
	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		//choke <- [2]string{msg.Topic(), string(msg.Payload())}
		msngr.Subscription <- string(msg.Payload())//+"\n"
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

	client.Disconnect(250)
	fmt.Println("Sample Subscriber Disconnected")
}

func Publish(payload string) {
	fmt.Println("Sample Publisher Started")
	opts := MQTT.NewClientOptions()
	opts.AddBroker(Config.MqttBroker)
	opts.SetClientID(Config.Name + "Pub")
	//opts.SetCleanSession(true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	num := 1
	qos := 0
	//payload := "This is a test and I'll see if it works."
	for i := 0; i < num; i++ {
		//fmt.Println("---- waiting for data to publish ----")
		//fmt.Println("---- doing publish ----", payload)
		token := client.Publish(Config.OutputTopic, byte(qos), false, payload)
		token.Wait()
	}

	client.Disconnect(250)
	fmt.Println("Sample Publisher Disconnected")

}

func publish() {
	for {
		//fmt.Println("hello")
		payload := <- msngr.Publication
		fmt.Println("PUBLISHING:", payload)
		Publish(payload)
		//fmt.Println("byebye")
	}
}