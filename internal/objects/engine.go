package objects

import (
	"fmt"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
	"github.com/eclipse/paho.mqtt.golang"
)

var (
	DefaultScript = `input -> output`
	DefaultInputDeviceNames = []string{}
	DefaultInputValueDescriptorNames = []string{}
	DefaultActuation = Actuation{}
	DefaultOffloadCondition = `func offload() string { return "local" })`
)

type Engine struct {
	Index                     int64       // index of engine, e.g. 1
	SinkSubscriber            mqtt.Client // mqtt client that connects from events of engine, e.g. "bitflow/engine/1/sink"
	ReverseCommandSubscriber  mqtt.Client // mqtt client that connects to reverse commands from engine, e.g. "bitflow/engine/1/reverse-command"
	Name                      string      // name of engine, e.g. engine-1
	Script                    string      // script contents
	InputDeviceNames          []string    // list of devices to get events from
	InputValueDescriptorNames []string    // list of value descriptors an event's readings are allowed to contain
	Actuation                 Actuation   // name of actuation in EdgeX
	Rule                      Rule        // rule based on actuation
	OffloadCondition          string      // offload function's signature and body
}

// sets index and adjusts name accordingly
func (e *Engine) setIndex(index int64) {
	e.Index = index
	e.Name = naming.Name(index)
}

// sets name and adjusts index accordingly
func (e *Engine) setName(name string) error {
	index, err := naming.ExtractIndex(name, "-", 1)
	if err != nil {
		err = fmt.Errorf("couldn't get index from device name: %v", err)
		return err
	}
	e.Name = name
	e.Index = index
	return nil
}

// set rule based on actuation
func (e *Engine) setRule(actuation Actuation) error {
	rule, err := actuation.InferRule()
	if err != nil {
		err = fmt.Errorf("couldn't get index from device name: %v", err)
		return err
	}
	e.Rule = rule
	return nil
}

func (e *Engine) Start() error {
	if e.Actuation.DeviceName != "" {
		rule, err := e.Actuation.InferRule()
		e.Rule = rule
		if err != nil {
			return fmt.Errorf("can't start engine %s, rule is faulty, actuation was %v", e.Name, e.Actuation)
		}
		err = rule.Add()
		if err != nil {
			return fmt.Errorf("can't start engine %s, rule couldn't be added to rules engine", e.Name)
		}
	}

	exportClient := ExportClient{
		BrokerName:   config.Broker.Name,
		BrokerSchema: config.Broker.Schema,
		BrokerHost:   config.Broker.Host,
		BrokerPort:   config.Broker.Port,
	}

	// add engine to export client
	exportClient.Add(*e)

	//sinkChannel := make(chan models.Event)
	//reverseCommandChannel := make(chan reverseCommandRequest)
	//go InitSinkSubscription(index, sinkChannel)
	//go InitReverseCommandSubscription(index, reverseCommandChannel)

	// TODO move handleSinkEvent here
	// TODO move handleReverseCommand here
	// TODO move InitRegistrySubscription here
	// TODO move InitSinkSubscription here

	//go createEngineInstance(e)
	return nil
}

// stop engine and clean up related values
func (e *Engine) Stop() error {
	// instruct engine to shutdown
	shutdown := "shutdown"
	communication.Publish(
		naming.Topic(e.Index, naming.Command),
		naming.Publisher(e.Index, naming.Command),
		shutdown)

	// stop subscribing to engine's commands and events
	communication.Disconnect(e.SinkSubscriber)
	communication.Disconnect(e.ReverseCommandSubscriber)

	// clean up rule, registration and device from EdgeX
	err := e.Rule.Remove()
	if err != nil {
		return fmt.Errorf("couldn't remove rule %s of engine %s from EdgeX: %v",
			e.Rule.Name, e.Name, err)
	}
	return nil

	// value descriptors are cleaned up in the background on behalf of device in
	// handleReverseCommand, when engine shuts down

	// TODO create graphic for MQTT hierarchy, whos's publishing what to whom and why
	// TODO explain MQTT hierarchy
	// remove registration
	// remove device from edgex
	// remove engine's index from registry
	// auto unsubscribe from channels...
	// unsubscribe "bitflow/engine/index_int/sink" as bitflow-engine-1-sink-subscriber
	// unsubscribe "bitflow/engine/index_int/reverse-command" as bitflow-engine-1-reverse-command-subscriber
}

// CLI publish to bitflow/engine/0/registry-request with text `0`
// CLI subscribe to bitflow/engine/0/registry-response
// if `failure` abort and report user engine already exists
// if `success` switch to API of Device Service

// CREATE ENGINE-N
// create object in engine-n
// save device name in engine-n object in engines map
// save device: (derive device -> body)
// - create device: POST to http://edgex-core-metadata:48081/api/v1/device
// DS on message to bitflow/engine/+/registry
//
// expect number `n` that is not already used, if it is already used
// if already used
// report `failure` back via publish to bitflow/engine/n/registry-response
// else
// create engine in engines map
// report `success` back via publish to bitflow/engine/n/registry-response


// await calls...
// WRITE ENGINE-N
// on call of DS command (script)
// on call of DS command (offload)
// on call of DS command (source)
// on call of DS command (actuation)
// write values in corresponding variable for engine-n object in engines map


// READ ENGINE-N and WRITE ENGINE-N
// on call of DS command (control=start)
// if source is []
// - return error, and let CLI notify user why this happened
// save export client name (derive export client -> body)
// - create export client: POST to http://edgex-export-client:48071/api/v1/registration
// if actuation is set (derive rule -> body)
// - create rule: POST to http://edgex-support-rulesengine:48075/api/v1/rule
// - subscribe to "bitflow/engine/n/sink"
// - subscribe to "bitflow/engine/n/reverse-command"

// ENGINE-N on message to bitflow/engine/0/reverse-command:
// - create value descriptors on behalf of a device: POST to
// POST http://edgex-core-data:48080/api/v1/valuedescriptor
// - publish response to "bitflow/engine/n/reverse-command-response

// ENGINE-N on message on "bitflow/engine/n/sink"
// POST http://edgex-core-data:48080/api/v1/event

// on call of DS command (control=shutdown)
// in ENGINE-N routine
// DELETE http://localhost:48081/api/v1/device/name/DEVICE_NAME
// DELETE http://localhost:48071/api/v1/registration/name/BitflowEngineNSourceTopic
// DELETE http://localhost:48075/api/v1/rule/name/RULE_NAME
// unsubscribe from "bitflow/engine/n/sink"
// unsubscribe from "bitflow/engine/n/reverse-command"
// in DS
// remove engine's index from registry
// and device from EdgeX

// DS end
// close registry channel
// unsubscribe from "bitflow/engine/0/registry-response"

// search through todos, especially cleanup