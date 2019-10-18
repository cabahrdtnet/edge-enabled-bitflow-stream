package objects

import (
	"fmt"
	"github.com/datenente/device-bitflow-stream/internal/communication"
	"github.com/datenente/device-bitflow-stream/internal/naming"
)

type Engine struct {
	booted                    bool        // is set when engine is started
	Index                     int64       // index of engine, e.g. 1
	Name                      string      // name of engine, e.g. engine-1
	Configuration             EngineConfiguration // configuration for current engine
	Communication             EngineCommunication
	Rule                      Rule        // rule based on actuation
	OffloadTarget			  string      // engine is offloaded to string target
}

func (e *Engine) HasBooted() bool {
	return e.booted
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

// start engine
func (e *Engine) start() error {
	e.Communication.Setup()

	instance := Instance{Engine: *e}
	err := instance.Create()
	if err != nil {
		return fmt.Errorf("couldn't start engine: %v", err)
	}

	e.booted = true
	return nil
}

// stop engine and deregister value descriptors
func (e *Engine) stop() error {
	e.booted = false

	// instruct engine to shutdown
	shutdown := "shutdown"
	communication.Publish(
		naming.Topic(e.Index, naming.Command),
		naming.Publisher(e.Index, naming.Command),
		shutdown)

	// stop subscribing to engine's commands and events
	e.Communication.Teardown()

	// TODO create graphic for MQTT hierarchy, whos's publishing what to whom and why
	// TODO explain MQTT hierarchy
	return nil
}

// set rule based on actuation
func (e *Engine) inferRule() error {
	rule, err := e.Configuration.InferRule()
	if err != nil {
		err = fmt.Errorf("couldn't infer rule for actuation %s: %v", e.Configuration.Actuation, err)
		return err
	}
	e.Rule = rule
	return nil
}

// check if engine can be started
// it can't started if it has been booted and or if it's misconfigured
func (e *Engine) startable() bool {
	if e.Configuration.InputDeviceNamesUnset() && e.Configuration.InputValueDescriptorNamesUnset() {
		return false
	}

	if e.booted {
		return false
	}

	return true
}

// add rule and register as export client
func (e *Engine) register() error {
	// add rule to rules engine
	if e.Configuration.ActuationSet() {
		err := e.inferRule()
		if err != nil {
			return fmt.Errorf("can't register engine %s as rule is faulty with actuation %v", e.Name, e.Configuration.Actuation)
		}
		err = e.Rule.Add()
		if err != nil {
			return fmt.Errorf("can't register engine %s as rule couldn't be added to rules engine", e.Name)
		}
	}

	// add engine as export client
	exportClient := ExportClient{}
	err := exportClient.Add(*e)
	if err != nil {
		return fmt.Errorf("can't register engine %s, rule couldn't be added to rules engine", e.Name)
	}

	// value descriptors init are automatically created by engine and sent to registry for registration
	return nil
}

// remove rule and export client registration but ignore value descriptors
// as a) it is unclear if other devices are also using these value descriptors
// and b) readings are never erased from by device-bitflow-stream, so there will always be data integrity
//        issues when attempting to erase a value descriptor
func (e *Engine) deregister() error {
	// remove rule associated with engine
	err := e.Rule.Remove()
	if err != nil {
		return fmt.Errorf("couldn't remove rule %s of engine %s: %v", e.Rule.Name, e.Name, err)
	}

	// remove engine as export client
	exportClient := ExportClient{}
	err = exportClient.Remove(*e)
	if err != nil {
		return fmt.Errorf("couldn't remove export client registration for engine %s: %v", e.Name, err)
	}

	// instruct engine to deregister value descriptors
	//deregister := "deregister"
	//communication.Publish(
	//	naming.Topic(e.Index, naming.Command),
	//	naming.Publisher(e.Index, naming.Command),
	//	deregister)

	// wait until value descriptors are deregistered

	return nil
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