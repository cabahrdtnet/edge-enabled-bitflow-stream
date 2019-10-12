package driver

import (
	"context"
	"fmt"
	"github.com/datenente/device-bitflow/internal/config"
	iModels "github.com/datenente/device-bitflow/internal/models"
	"github.com/datenente/device-bitflow/internal/naming"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"strconv"
)

func startEngine(name string) error {
	BitflowDriver.mutex.RLock()
	engine, exists := BitflowDriver.engines[name]
	if ! exists {
		return fmt.Errorf("can't start engine %s does not exist", engine.Name)
	}
	BitflowDriver.mutex.RUnlock()

	index, err := naming.ExtractIndex(engine.Name, "-", 1)
	if err != nil {
		return fmt.Errorf("can't start engine %s, no index found", engine.Name)
	}

	if engine.Actuation.DeviceName != "" {
		rule, err := iModels.From(engine.Actuation, index)
		if err != nil {
			return fmt.Errorf("can't start engine %s, rule is faulty, actuation was %v", engine.Name, engine.Actuation)
		}
		err = rule.Add()
		if err != nil {
			return fmt.Errorf("can't start engine %s, rule couldn't be added to rules engine", engine.Name)
		}
	}

	brokerPort64, err := strconv.ParseInt(BitflowDriver.config.BrokerPort, 10, 64)
	if err != nil {
		return fmt.Errorf("couldn't parse broker port in because of: %v", err)
	}
	brokerPort := int(brokerPort64)

	registration := models.Registration{
		ID:          "",
		Created:     0,
		Modified:    0,
		Origin:      0,
		Name:        naming.ExportName(index, "source"),
		Addressable: models.Addressable{
			Timestamps: models.Timestamps{},
			Id:         "",
			Name:       "MosquittoBroker",
			Protocol:   BitflowDriver.config.BrokerSchema,
			HTTPMethod: "",
			Address:    BitflowDriver.config.BrokerHost,
			Port:       brokerPort,
			Path:       "",
			Publisher:  naming.Publisher(index, naming.Source),
			User:       "",
			Password:   "",
			Topic:      naming.Topic(index, naming.Source),
		},
		Format:      "JSON",
		Filter:      models.Filter{
			DeviceIDs:          engine.InputDeviceNames,
			ValueDescriptorIDs: engine.InputValueDescriptorNames,
		},
		Encryption:  models.EncryptionDetails{
			Algo:       "",
			Key:        "",
			InitVector: "",
		},
		Compression: "",
		Enable:      true,
		Destination: "MQTT_TOPIC",
	}

	url := config.URL.ExportClient + clients.ApiRegistrationRoute
	_, err = clients.PostJsonRequest(url, registration, context.TODO())
	if err != nil {
		return fmt.Errorf("couldn't send rule to rules engine: %v", err)
	}
	sinkChannel := make(chan models.Event)
	reverseCommandChannel := make(chan reverseCommandRequest)
	go InitSinkSubscription(index, sinkChannel)
	go InitReverseCommandSubscription(index, reverseCommandChannel)
	go createEngineInstance(index)
	return nil
}

// TODO implement this
func createEngineInstance(index int64) {
	// -name="engine-0"
	// -script="input -> avg() -> append_latency() -> output"
	// -input="bitflow/engine/0/source"
	// -output="bitflow/engine/0/sink"
	// -command="bitflow/engine/0/command"
	// -reverse-command="bitflow/engine/0/reverse-command"
	// -reverse-command-response="bitflow/engine/0/reverse-command-response"
	// -broker="tcp://192.168.178.20:1883"
	// -bitflow-params="-v -buf 20000"
	//
	// offload? interpret and run local or remote
	// just set different DOCKER_HOST from remoteDockerHostSchema/Host/Port config
	//
	// docker run
}

// TODO implement this
// stop engine and clean up related values
func stopEngine(name string) {
	// send shutdown to engine
	// let engine clean up value descriptors
	// remove rule
	// remove registration
	// remove device from edgex
	// remove engine's index from registry
	// auto unsubscribe from channels...
}