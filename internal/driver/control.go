package driver

import (
	"encoding/json"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
)

type response struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// TODO move all this into enginecontrol

func InitRegistrySubscription() {
	config.Log.Debug("Subscribing to " + naming.RegistryRequest)
	subscriber.registry = communication.Subscribe(
		naming.Topic(-1, naming.RegistryRequest),
		naming.Subscriber(-1, naming.RegistryRequest),
		handleRegistryRequestMessage)
}

// per device subscription for sink of engine
func handleRegistryRequest() {
	for {
		select {
		case request := <- registration.request:
			var rsp response
			var err error
			var index int64

			if request > 0 {
				index = request
				err = registry.Register(index)
			} else {
				index = -request
				err = registry.Deregister(index)
			}

			if err == nil {
				rsp.Message = "success"
				rsp.Error = ""
			} else {
				rsp.Message = "failure"
				rsp.Error = err.Error()
			}

			payload, _ := json.Marshal(rsp)
			topic := naming.Topic(index, naming.RegistryResponse)
			clientID := naming.Publisher(index, naming.RegistryResponse)
			msg := string(payload)
			communication.Publish(topic, clientID, msg)

		case err := <- registration.err:
			var rsp response

			rsp.Message = "failure"
			rsp.Error = err.Error()

			payload, err := json.Marshal(rsp)
			if err != nil {
				log.Error(err.Error())
			}
			msg := string(payload)
			log.Error(msg)
			// TODO send this back to server
		}
	}
}