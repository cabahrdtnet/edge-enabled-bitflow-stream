package driver

import (
	"encoding/json"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/naming"
)

type response struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func InitRegistrySubscription() {
	driver.lc.Debug("Subscribing to " + naming.RegistryRequest)
	subscriber.registry = communication.Subscribe(
		naming.Topic(-1, naming.RegistryRequest),
		naming.Subscriber(-1, naming.RegistryRequest),
		handleRegistryRequestMessage)
}

func handleRegistryRequest() {
	for {
		select {
		case request := <- registration.request:
			var rsp response
			var err error
			var index int64

			if request > 0 {
				index = request
				err = register(index)
			} else {
				index = -request
				err = deregister(index)
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
				driver.lc.Error(err.Error())
			}
			msg := string(payload)
			driver.lc.Error(msg)
			// publish this back to client
		}
	}
	driver.lc.Info("registry.request is closed.")
}