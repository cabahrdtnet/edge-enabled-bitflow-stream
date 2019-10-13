package objects

import (
	"github.com/google/go-cmp/cmp"
)

var (
	DefaultScript = `input -> output`
	DefaultInputDeviceNames = []string{}
	DefaultInputValueDescriptorNames = []string{}
	DefaultActuation = Actuation{}
	DefaultOffloadCondition = `func offload() string { return "local" })`
)

type EngineConfiguration struct {
	Script                    string    // script contents
	InputDeviceNames          []string  // list of devices to get events from
	InputValueDescriptorNames []string  // list of value descriptors an event's readings are allowed to contain
	Actuation                 Actuation // name of actuation in EdgeX
	OffloadCondition          string    // offload function's signature and body
}

// infer rule based on actuation in configuration
func (ec* EngineConfiguration) InferRule() (Rule, error) {
	return ec.Actuation.inferRule()
}

// check if script is set
func (ec *EngineConfiguration) ScriptSet() bool {
	return ! ec.ScriptUnset()
}

// check if script is unset
func (ec *EngineConfiguration) ScriptUnset() bool {
	if ec.Script == DefaultScript || ec.Script == "" {
		return true
	} else {
		return false
	}
}

// check if input device names are set
func (ec *EngineConfiguration) InputDeviceNamesSet() bool {
	return ! ec.InputDeviceNamesUnset()
}

// check if input device names are unset
func (ec *EngineConfiguration) InputDeviceNamesUnset() bool {
	if len(ec.InputDeviceNames) == 0 {
		return true
	} else {
		return false
	}
}

// check if input value descriptor names are set
func (ec *EngineConfiguration) InputValueDescriptorNamesSet() bool {
	return ! ec.InputValueDescriptorNamesUnset()
}

// check if input value descriptor names are unset
func (ec *EngineConfiguration) InputValueDescriptorNamesUnset() bool {
	if len(ec.InputValueDescriptorNames) == 0 {
		return true
	} else {
		return false
	}
}

// check if actuation is set
func (ec *EngineConfiguration) ActuationSet() bool {
	return ! ec.ActuationUnset()
}

// check if actuation is unset
func (ec *EngineConfiguration) ActuationUnset() bool {
	return cmp.Equal(ec.Actuation, DefaultActuation)
}

// check if offloading condition is set
func (ec *EngineConfiguration) OffloadingConditionSet() bool {
	return ! ec.OffloadingConditionUnset()
}

// check if offloading condition is unset
func (ec *EngineConfiguration) OffloadingConditionUnset() bool {
	if ec.OffloadCondition == DefaultOffloadCondition || ec.OffloadCondition == "" {
		return true
	} else {
		return false
	}
}

