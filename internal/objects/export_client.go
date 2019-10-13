package objects

import (
	"context"
	"fmt"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type ExportClient struct {
}

// add engine as export client
func (ec *ExportClient) Add(engine Engine) error {
	registration := models.Registration{
		ID:          "",
		Created:     0,
		Modified:    0,
		Origin:      0,
		Name:        naming.ExportName(engine.Index, "source"),
		Addressable: models.Addressable{
			Timestamps: models.Timestamps{},
			Id:         "",
			Name:       "MosquittoBroker",
			Protocol:   config.Broker.Schema,
			HTTPMethod: "",
			Address:    config.Broker.Host,
			Port:       config.Broker.Port,
			Path:       "",
			Publisher:  naming.Publisher(engine.Index, naming.Source),
			User:       "",
			Password:   "",
			Topic:      naming.Topic(engine.Index, naming.Source),
		},
		Format:      "JSON",
		Filter:      models.Filter{
			DeviceIDs:          engine.Configuration.InputDeviceNames,
			ValueDescriptorIDs: engine.Configuration.InputValueDescriptorNames,
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

	err := requestPost(registration, engine.Index)
	if err != nil && err.(ExportClientError).OtherError {
		return fmt.Errorf("can't add export client: %v", err)
	}

	return nil
}

// remove engine from as export client
func (ec * ExportClient) Remove(engine Engine) error {
	err := requestDeletion(engine.Index)
	if err != nil {
		fmt.Errorf("couldn't remove engine with name %s: %v", engine.Name, err)
	}
	return nil
}

// post registration
func requestPost(registration models.Registration, index int64) error {
	engineName := naming.Name(index)
	url := config.URL.ExportClient + clients.ApiRegistrationRoute
	_, err := clients.PostJsonRequest(url, registration, context.TODO())
	if err != nil && err.(*types.ErrServiceClient).StatusCode == 400 {
		return ExportClientError{
			EngineName: engineName,
			AlreadyRegistered: true,
		}
	}
	if err != nil && err.(*types.ErrServiceClient).StatusCode != 400 {
		return ExportClientError{
			EngineName: engineName,
			OtherError: true,
		}
	}
	return nil
}

// remove registration by name
func requestDeletion(index int64) error {
	engineName := naming.Name(index)
	name := naming.ExportName(index, "source")
	url := config.URL.ExportClient + clients.ApiRegistrationRoute + "/name/" + name
	err := clients.DeleteRequest(url, context.TODO())
	if err != nil && err.(*types.ErrServiceClient).StatusCode == 400 {
		return ExportClientError{
			EngineName: engineName,
			AlreadyRegistered: true,
		}
	}
	return nil
}
