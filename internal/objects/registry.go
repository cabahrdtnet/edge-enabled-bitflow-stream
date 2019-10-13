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
		r.addIndexedDefaultEngine(index)
	}

	return nil
}

// register and start engine
func (r *Registry) Start(index int64) error {
	engine, err := r.get(index)
	if err != nil {
		return fmt.Errorf("couldn't start engine as engine couldn't be retrieved: %v", err)
	}

	err = engine.register()
	if err != nil {
		return fmt.Errorf("couldn't start engine as engine couldn't register: %v", err)
	}

	if ! engine.startable() {
		return fmt.Errorf("couldn't start engine as engine isn't startable")
	}

	err = engine.start()
	if err != nil {
		return fmt.Errorf("couldn't start engine as engine couldn't start: %v", err)
	}

	return nil
}

// TODO write prose about this
// when what is registered and what happens when engine is started and stopped
// device, rule, registration, value descriptors

// stop engine
func (r *Registry) Stop(index int64) error {
	engine, err := r.get(index)
	if err != nil {
		return fmt.Errorf("couldn't stop engine: %v", err)
	}

	err = engine.stop()
	if err != nil {
		return fmt.Errorf("couldn't start engine: %v", err)
	}

	return nil
}

// set engine identified by index with template engine
func (r *Registry) Update(index int64, template Engine) error {
	name := naming.Name(index)
	engine, err := r.get(index)
	if err != nil {
		return fmt.Errorf("can't update engine with name %s, because it does not exist in engine registry", name)
	}

	if template.Configuration.ScriptSet() {
		engine.Configuration.Script = template.Configuration.Script
	}

	if template.Configuration.InputDeviceNamesSet() {
		engine.Configuration.InputDeviceNames = append(
				template.Configuration.InputDeviceNames[:0:0],
				template.Configuration.InputDeviceNames...)
	}

	if template.Configuration.InputValueDescriptorNamesSet() {
		engine.Configuration.InputValueDescriptorNames = append(
			template.Configuration.InputValueDescriptorNames[:0:0],
			template.Configuration.InputValueDescriptorNames...)
	}

	if template.Configuration.ActuationSet() {
		engine.Configuration.Actuation = Actuation{
			DeviceName:   template.Configuration.Actuation.DeviceName,
			CommandName:  template.Configuration.Actuation.CommandName,
			CommandBody:  template.Configuration.Actuation.CommandBody,
			LeftOperand:  template.Configuration.Actuation.LeftOperand,
			Operator:     template.Configuration.Actuation.Operator,
			RightOperand: template.Configuration.Actuation.RightOperand,
		}
	}

	if template.Configuration.OffloadingConditionSet() {
		engine.Configuration.OffloadCondition = template.Configuration.OffloadCondition
	}

	r.engines[name] = engine
	return nil
}

// add engine to engine registry and add device
func (r *Registry) Register(index int64) error {
	err := r.addIndexedDefaultEngine(index)
	if err != nil {
		return fmt.Errorf("couldn't register engine in engine registry: %v", err)
	}

	err = r.addDevice(index)
	if err != nil {
		return fmt.Errorf("couldn't register engine in engine registry: %v", err)
	}

	return nil
}

// deregister all associated metadata of a stopped engine
func (r *Registry) Deregister(index int64) error {
	name := naming.Name(index)
	engine, err  := r.get(index)
	if err != nil {
		return fmt.Errorf("couldn't deregister %s couldn't be retrieved: %v", name, err)
	}

	// check if engine still runs, if so stop
	if engine.HasBooted() {
		err = engine.stop()
		if err != nil {
			return fmt.Errorf("couldn't deregister %s, as %s couldn't be stopped: %v", name, name, err)
		}
	}

	// remove device associated with engine
	err = r.removeDevice(index)
	if err != nil {
		return fmt.Errorf("couldn't deregister %s as device couldn't be removed: %v", name, err)
	}

	// remove rule and registration in engine
	err = engine.deregister()
	if err != nil {
		return fmt.Errorf("couldn't deregister %s as engine couldn't deregister: %v", name, err)
	}

	// finally delete engine from registry
	err = r.deleteEngine(index)
	if err != nil {
		return fmt.Errorf("couldn't deregister %s as engine couldn't be deleted from engine registry: %v", name, err)
	}

	return nil
}

// get engine by index
func (r *Registry) get(index int64) (Engine, error) {
	r.mutex.RLock()
	defer r.mutex.Unlock()
	name := naming.Name(index)
	engine, exists := r.engines[name]
	if ! exists {
		return Engine{}, fmt.Errorf("can't get engine: %s does not exist", engine.Name)
	}
	return engine, nil
}

// add engine to engine registry
func (r *Registry) addIndexedDefaultEngine(index int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := naming.Name(index)
	_, exists := r.engines[name]
	if ! exists {
		r.engines[name] = Engine{
			Index: index,
			Name:  name,
			Configuration: EngineConfiguration{
				Script:                    DefaultScript,
				InputDeviceNames:          DefaultInputDeviceNames,
				InputValueDescriptorNames: DefaultInputValueDescriptorNames,
				Actuation:                 Actuation{},
				OffloadCondition:          DefaultOffloadCondition,
			},
			Communication: EngineCommunication{
				Index:                    index,
				SinkSubscriber:           nil,
				ReverseCommandSubscriber: nil,
				Events:                   make(chan contract.Event),
				ReverseCommandRequests:   make(chan reverseCommandRequest),
			},
			Rule:          Rule{},
			OffloadTarget: naming.Local,
		}
		return nil
	} else {
		return fmt.Errorf("can't add engine with name %s, because it already exists in engine registry", name)
	}
}

// delete engine from engine registry
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

// add device associated with engine
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

// remove engine's associated device
func (r *Registry) removeDevice(index int64) error {
	name := naming.Name(index)
	err := sdk.RunningService().RemoveDeviceByName(name)
	if err != nil {
		return fmt.Errorf("couldn't remove device from EdgeX: %v", err)
	}
	return nil
}