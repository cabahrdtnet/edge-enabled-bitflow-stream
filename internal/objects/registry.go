package objects

import (
	"fmt"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
	sdk "github.com/edgexfoundry/device-sdk-go"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"sync"
)

type Registry struct {
	engines map[string]Engine
	mutex   sync.RWMutex
}

// as the registry is not persisted, we dynamically load it from existing devices
func (r *Registry) Init() error {
	r.engines = make(map[string]Engine)
	devices := sdk.RunningService().Devices()

	for _, device := range devices {
		index, err := naming.ExtractIndex(device.Name, "-", 1)
		if err != nil {
			return fmt.Errorf("could not add engine to registry: %v", err)
		}
		r.addEngine(index)
	}

	return nil
}

// get engine by name
func (r *Registry) Get(name string) (Engine, error) {
	r.mutex.RLock()
	defer r.mutex.Unlock()
	engine, exists := r.engines[name]
	if ! exists {
		return Engine{}, fmt.Errorf("can't start engine %s does not exist", engine.Name)
	}
	return engine, nil
}

// set engine identified by index with template
func (r *Registry) Update(index int64, template Engine) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	name := naming.Name(index)
	engine, exists := r.engines[name]
	if exists {
		if template.Script != "" {
			engine.Script = template.Script
		}

		if len(template.InputDeviceNames) != 0 {
			engine.InputDeviceNames =
				append(template.InputDeviceNames[:0:0], template.InputDeviceNames...)
		}

		if len(template.InputValueDescriptorNames) != 0 {
			engine.InputValueDescriptorNames =
				append(template.InputValueDescriptorNames[:0:0], template.InputValueDescriptorNames...)
		}

		if template.Actuation.DeviceName != "" {
			engine.Actuation.DeviceName = template.Actuation.DeviceName
			engine.Actuation.CommandName = template.Actuation.CommandName
			engine.Actuation.CommandBody = template.Actuation.CommandBody
			engine.Actuation.LeftOperand = template.Actuation.LeftOperand
			engine.Actuation.Operator = template.Actuation.Operator
			engine.Actuation.RightOperand = template.Actuation.RightOperand
		}

		if template.OffloadCondition != "" {
			engine.OffloadCondition = template.OffloadCondition
		}

		r.engines[name] = engine
		return nil
	} else {
		return fmt.Errorf("can't change engine with name %s, because it does not exist in engine registry", name)
	}
}

// add engine based on index to engine registry and add associated device to EdgeX
func (r *Registry) Register(index int64) error {
	err := r.addEngine(index)
	if err != nil {
		return fmt.Errorf("couldn't register engine in engine registry: %v", err)
	}

	err = r.addDevice(index)
	if err != nil {
		return fmt.Errorf("couldn't register engine in engine registry: %v", err)
	}

	return nil
}

// remove engine based on index from engine registry and remove device from EdgeX
func (r *Registry) Deregister(index int64) error {
	err := r.deleteEngine(index)
	if err != nil {
		return err
	}

	err = r.removeDevice(index)
	if err != nil {
		return err
	}

	return nil
}

// add engine to engine registry
func (r *Registry) addEngine(index int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := naming.Name(index)
	_, exists := r.engines[name]
	if ! exists {
		r.engines[name] = Engine{
			Name:                      name,
			Script:                    DefaultScript,
			InputDeviceNames:          DefaultInputDeviceNames,
			InputValueDescriptorNames: DefaultInputValueDescriptorNames,
			Actuation:                 Actuation{},
			OffloadCondition:          DefaultOffloadCondition,
		}
		return nil
	} else {
		return fmt.Errorf("can't add engine with name %s, because it already exists in engine registry", name)
	}
}

// delete engine from engine registry in r.Driver
func (r *Registry) deleteEngine(index int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := naming.Name(index)
	_, exists := r.engines[name]

	if exists {
		delete(r.engines, name)
		return nil
	} else {
		return fmt.Errorf("can't delete engine with name %s, because it does not exist in engine registry", name)
	}
}

// add engine as device to EdgeX
func (r *Registry) addDevice(index int64) error {
	name := naming.Name(index)
	props := contract.ProtocolProperties{
		"ClientId" : naming.Publisher(index, naming.Command),
		"Host" :     config.Broker.Host,
		"Password" : "",
		"Port" :     fmt.Sprintf("%d", config.Broker.Port),
		"Schema" :   config.Broker.Schema,
		"Topic" :    naming.Topic(index, naming.Command),
		"User" :     "",
	}

	dev := contract.Device{
		DescribedObject: contract.DescribedObject{},
		Id:              "",
		Name:            name,
		AdminState:      "unlocked",
		OperatingState:  "enabled",
		Protocols: map[string]contract.ProtocolProperties{
			naming.Protocol: props,
		},
		LastConnected: 0,
		LastReported:  0,
		Labels:        []string{
			"bitflow-script-execution-engine",
			"created-by-device-bitflow",
		},
		Location:        nil,
		Service:         contract.DeviceService{
			Name: naming.DeviceService,
			Addressable: contract.Addressable{
					Name: naming.DeviceService,
			},
		},
		Profile:         contract.DeviceProfile{
			Name: naming.DeviceProfile,
		},
		AutoEvents:      nil,
	}

	_, err := sdk.RunningService().AddDevice(dev)
	if err != nil {
		return fmt.Errorf("couldn't add device to EdgeX: %v", err)
	}
	return err
}

// remove engine's associated device in EdgeX
func (r *Registry) removeDevice(index int64) error {
	name := naming.Name(index)
	err := sdk.RunningService().RemoveDeviceByName(name)
	if err != nil {
		return fmt.Errorf("couldn't remove device from EdgeX: %v", err)
	}
	return nil
}