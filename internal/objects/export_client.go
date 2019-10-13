package objects

import (
	"context"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type ExportClient struct {
	BrokerName string
	BrokerSchema string
	BrokerHost string
	BrokerPort int
}

func (r *ExportClient) Add(engine Engine) error {
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
			Protocol:   r.BrokerSchema,
			HTTPMethod: "",
			Address:    r.BrokerHost,
			Port:       r.BrokerPort,
			Path:       "",
			Publisher:  naming.Publisher(engine.Index, naming.Source),
			User:       "",
			Password:   "",
			Topic:      naming.Topic(engine.Index, naming.Source),
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

	err := r.post(registration, engine.Name)
	return err
}

func (r *ExportClient) post(registration models.Registration, engineName string) error {
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

