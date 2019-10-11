package models

var (
	DefaultScript = `input -> output`
	DefaultInputDeviceNames = []string{}
	DefaultInputValueDescriptorNames = []string{}
	DefaultActuation = Actuation{}
	DefaultOffloadCondition = `func offload() string { return "local" })`
)

type Engine struct {
	Name string // name of engine, e.g. engine-0
	Script string // script contents
	InputDeviceNames []string // list of devices to get events from
	InputValueDescriptorNames []string // list of value descriptors an event's readings are allowed to contain
	Actuation Actuation // name of actuation in EdgeX
	OffloadCondition string // offload function's signature and body
}

type Actuation struct {
	DeviceName string
	CommandName string
	CommandBody string
	LeftOperand string
	Operator string
	RightOperand string
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
// save export client name (derive export client -> body
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

// remove device metadata only if device is started and stopped by device CLI
