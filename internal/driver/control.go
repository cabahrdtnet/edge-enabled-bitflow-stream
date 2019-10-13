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
				err = driver.registry.Register(index)
			} else {
				index = -request
				err = driver.registry.Deregister(index)
			}

			if err == nil {
				rsp.Message = "success"
				rsp.Error = ""
			} else {
				rsp.Message = "failure"
				rsp.Error = err.Error()
			}

			publishResponse(rsp, index)

		case err := <- registration.err:
			var rsp response

			rsp.Message = "failure"
			rsp.Error = err.value.Error()

			payload, payloadErr := json.Marshal(rsp)
			if payloadErr != nil {
				config.Log.Error(payloadErr.Error())
			}
			msg := string(payload)
			config.Log.Error(msg)
			publishResponse(rsp, err.index)
		}
	}
}

// publish response to engine with given index
func publishResponse(rsp response, index int64) {
	payload, _ := json.Marshal(rsp)
	topic := naming.Topic(index, naming.RegistryResponse)
	clientID := naming.Publisher(index, naming.RegistryResponse)
	msg := string(payload)
	communication.Publish(topic, clientID, msg)
}