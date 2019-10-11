package driver

import (
	"fmt"
	"github.com/datenente/device-bitflow/internal/naming"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	registerIndex = "register"
	deregisterIndex = "deregister"
)

var (
	//events = struct {
	//	outgoing chan models.Event
	//	incoming chan models.Event
	//}{
	//	make(chan models.Event),
	//	make(chan models.Event)}
	//
	//initial = struct {
	//	outputHeader chan string
	//	processedEvent chan models.Event
	//}{
	//	make(chan string),
	//	make(chan models.Event)}
	//
	//commands = struct {
	//	incoming chan string
	//}{
	//	make(chan string)}
	//
	//// TODO move to outgoing commands of commands
	//reverseCommandResponse = struct {
	//	incoming chan string
	//}{
	//	make(chan string)}

	registration = struct{
		request chan int64
		err chan error
	}{
		make(chan int64),
		make(chan error)}

	subscriber = struct{
		registry MQTT.Client
	}{}
)

// handles registry message for devices so that device service API can be used on a device
func handleRegistryRequestMessage(client MQTT.Client, msg MQTT.Message) {
	index, err := naming.ExtractIndex(msg.Topic(), "/", 2)
	if err != nil {
		registration.err <- err
		return
	}

	if index == 0 {
		registration.err <- fmt.Errorf("engine indices start with 1, not 0")
		return
	}

	requestCommand := string(msg.Payload())
	switch requestCommand {
	case registerIndex:
		registration.request <- index

	case deregisterIndex:
		registration.request <- -index

	default:
		registration.err <- fmt.Errorf("wrong registration request command: allowed are register/deregister")
	}
}