package driver

import (
	"fmt"
	"github.com/datenente/device-bitflow/internal/models"
	"github.com/datenente/device-bitflow/internal/naming"
	sdk "github.com/edgexfoundry/device-sdk-go"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// as the registry is not persisted, we dynamically load it from existing devices
func initEngineRegistry() error {
	devices := sdk.RunningService().Devices()

	for _, device := range devices {
		index, err := naming.ExtractIndex(device.Name, "-", 1)
		if err != nil {
			return fmt.Errorf("could not add engine to registry: %v", err)
		}
		addEngine(index)
	}

	return nil
}

func update(index int64, template models.Engine) error {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()
	name := naming.Name(index)
	engine, exists := driver.engines[name]
	if exists {
		if engine.Name == "" {
			engine.Name = template.Name
		}

		if engine.Script == "" {
			engine.Script = template.Script
		}

		if len(engine.InputDeviceNames) == 0 {
			engine.InputDeviceNames =
				append(template.InputDeviceNames[:0:0], template.InputDeviceNames...)
		}

		if len(engine.InputValueDescriptorNames) == 0 {
			engine.InputValueDescriptorNames =
				append(template.InputValueDescriptorNames[:0:0], template.InputValueDescriptorNames...)
		}

		if engine.Actuation.DeviceName == "" {
			engine.Actuation.DeviceName = template.Actuation.DeviceName
			engine.Actuation.CommandName = template.Actuation.CommandName
			engine.Actuation.CommandBody = template.Actuation.CommandBody
			engine.Actuation.LeftOperand = template.Actuation.LeftOperand
			engine.Actuation.Operator = template.Actuation.Operator
			engine.Actuation.RightOperand = template.Actuation.RightOperand
		}

		if engine.OffloadCondition == "" {
			engine.OffloadCondition = template.OffloadCondition
		}

		driver.lc.Debug("updated engine " + naming.Name(index) + " in engine registry")
		return nil
	} else {
		return fmt.Errorf("can't change engine with name %s, because it does not exist in engine registry", name)
	}
}

// add engine to engine registry in driver and add associated device to EdgeX
func register(index int64) error {
	err := addEngine(index)
	if err != nil {
		return fmt.Errorf("couldn't register engine in engine registry: %v", err)
	}

	err = addDevice(index)
	if err != nil {
		return fmt.Errorf("couldn't register engine in engine registry: %v", err)
	}

	return nil
}

// remove engine from engine registry in driver and remove device from EdgeX
func deregister(index int64) error {
	err := deleteEngine(index)
	if err != nil {
		return err
	}

	err = removeDevice(index)
	if err != nil {
		return err
	}

	return nil
}

// TODO write unit test for this
// add engine to engine registry in driver
func addEngine(index int64) error {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	name := naming.Name(index)
	_, exists := driver.engines[name]
	if ! exists {
		driver.engines[name] = models.Engine{
			Name:                      name,
			Script:                    models.DefaultScript,
			InputDeviceNames:          models.DefaultInputDeviceNames,
			InputValueDescriptorNames: models.DefaultInputValueDescriptorNames,
			Actuation:                 models.Actuation{},
			OffloadCondition:          models.DefaultOffloadCondition,
		}
		driver.lc.Debug("added engine " + naming.Name(index) + " to engine registry")
		return nil
	} else {
		return fmt.Errorf("can't add engine with name %s, because it already exists in engine registry", name)
	}
}

// delete engine from engine registry in driver
func deleteEngine(index int64) error {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	name := naming.Name(index)
	_, exists := driver.engines[name]

	if exists {
		delete(driver.engines, name)
		driver.lc.Debug("deleted engine " + name + " from engine registry")
		return nil
	} else {
		return fmt.Errorf("can't delete engine with name %s, because it does not exist in engine registry", name)
	}
}

// add engine as device to EdgeX
func addDevice(index int64) error {
	name := naming.Name(index)
	props := contract.ProtocolProperties{
		"ClientId" : naming.Publisher(index, naming.Command),
		"Host" : driver.config.BrokerHost,
		"Password" : "",
		"Port" : driver.config.BrokerPort,
		"Schema" : driver.config.BrokerSchema,
		"Topic" : naming.Topic(index, naming.Command),
		"User" : "",
	}

	//url := fmt.Sprintf("http://%s:%d%s",
	//	clients.ExportClientServiceKey, 48071,
	//	clients.ApiRegistrationRoute)
	//url := fmt.Sprintf("http://%s:%d%s",
	//	"localhost", 48071,
	//	clients.ApiRegistrationRoute)
	//url := fmt.Sprintf("http://%s:%d%s",
	//	"localhost", 48081,
	//	clients.ApiDeviceRoute)
//http://edgex-core-metadata:48081/api/v1/device

	dev := contract.Device{
		DescribedObject: contract.DescribedObject{},
		Id:              "",
		Name:            name,
		AdminState:      "unlocked",
		OperatingState:  "enabled",
		Protocols: map[string]contract.ProtocolProperties{
			Protocol: props,
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
	//_, err := clients.PostJsonRequest(url, device, context.TODO())
	if err == nil {
		driver.lc.Debug("added device " + naming.Name(index) + " to EdgeX")
		return err
	} else {
		return fmt.Errorf("couldn't add device to EdgeX: %v", err)
	}
}

// remove engine's associated device in EdgeX
func removeDevice(index int64) error {
	name := naming.Name(index)
	err := sdk.RunningService().RemoveDeviceByName(name)
	if err == nil {
		driver.lc.Debug("removed device " + naming.Name(index) + " from EdgeX")
		return nil
	} else {
		return fmt.Errorf("couldn't remove device from EdgeX: %v", err)
	}
}