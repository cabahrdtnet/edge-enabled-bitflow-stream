package driver

import (
	"context"
	"fmt"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/config"
	iModels "github.com/datenente/device-bitflow/internal/models"
	"github.com/datenente/device-bitflow/internal/naming"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"io/ioutil"
	"os"
	"os/exec"
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
	go createEngineInstance(engine)
	return nil
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

func createEngineInstance(engine iModels.Engine) {
	defer unsetDocker()

	name := engine.Name
	index, err := naming.ExtractIndex(name, "-", 1)
	if err != nil {
		BitflowDriver.lc.Error("couldn't extract index for " + name + " in engine instance" )
		return
	}

	script := engine.Script
	input := naming.Topic(index, naming.Source)
	output := naming.Topic(index, naming.Sink)
	command := naming.Topic(index, naming.Command)
	reverseCommand := naming.Topic(index, naming.ReverseCommand)
	reverseCommandResponse := naming.Topic(index, naming.ReverseCommandResponse)
	broker := communication.Broker
	condition := engine.OffloadCondition
	//bitflowParams := engine.BitflowParams

	// offloading
	location, err := where(index, condition)
	if err != nil {
		formatted := fmt.Sprintf("couldn't derive offload location: %v", err)
		BitflowDriver.lc.Error(formatted)
		return
	}

	if location == "local" {
		setLocalDocker()
		return
	}

	if location == "remote" {
		setRemoteDocker()
		return
	}

	docker := "docker"
	image := config.DockerEngineImage
	args := []string{"run", image,
		"-name", name,
		"-script", script,
		"-input", input,
		"-output", output,
		"-command", command,
		"-reverseCommand", reverseCommand,
		"-reverseCommandResponse", reverseCommandResponse,
		"-broker", broker,
		//-"bitflow-params"="-v -buf 20000"
	}

	cmd := exec.Command(docker, args...)
	if err := cmd.Run(); err != nil {
		formatted := fmt.Sprintf("couldn't run docker instance with args %s: %v", args, err)
		BitflowDriver.lc.Error(formatted)
	}
}

// decide where offloading occurs, either locally or remotely
func where(index int64, condition string) (string, error) {
	fileName := naming.Name(index) + "-" + "offload.go"
	defer os.Remove(fileName)
	data := []byte(condition)
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return "", fmt.Errorf("couldn't write temp offload condition file %s: %v", fileName, err)
	}

	goBin := "go"
	args := []string{"run", fileName}
	cmd := exec.Command(goBin, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("couldn't write temp offload condition file %s: %v", fileName, err)
	}
	output := string(outputBytes)
	return output, nil
}

func set(key string, env string) error {
	err := os.Setenv(key, env)
	if err != nil {
		return fmt.Errorf("couldn't set %s env variable: %v", env, err)
	}
	return nil
}

func unset(key string) error {
	err := os.Unsetenv(key)
	if err != nil {
		return fmt.Errorf("couldn't unset %s env variable: %v", key, err)
	}
	return nil
}

func setLocalDocker() {
	set(config.DockerTLSVerify, config.Docker.LocalDockerTLSVerify)
	set(config.DockerHost, config.Docker.LocalDockerHost)
	set(config.DockerCertPath, config.Docker.LocalDockerCertPath)
	set(config.DockerMachineName, config.Docker.LocalDockerMachineName)
}

func setRemoteDocker() {
	set(config.DockerTLSVerify, config.Docker.RemoteDockerTLSVerify)
	set(config.DockerHost, config.Docker.RemoteDockerHost)
	set(config.DockerCertPath, config.Docker.RemoteDockerCertPath)
	set(config.DockerMachineName, config.Docker.RemoteDockerMachineName)
}

func unsetDocker() {
	unset(config.DockerTLSVerify)
	unset(config.DockerHost)
	unset(config.DockerCertPath)
	unset(config.DockerMachineName)
}