package engine

var (
	// these values are set in cmd/device-bitflow/engine/main.go
	Config = Configuration{}
	msngr = messenger{make(chan string), make(chan string)}
)

// apply configuration for a run
func Configure() {
	//go func() {
	//	for msg := range msngr.Subscription {
	//		fmt.Println("Received: ", msg)
	//	}
	//}()

	go publish()
	go InitSubscription()
}

// MQTT messages
// sub handler      writesTo  Message.Subscription
// stdin of bitflow readsFrom Message.Subscription

// processing
// stdout of bitflow writesTo  Message.Publication
// publish           readsFrom Message.Publication