package main

import (
	"flag"
	"fmt"
	"github.com/datenente/device-bitflow/internal/engine"
)

/*
- script
- input endpoint
- output endpoint
- mqtt broker authority
- engine name
 */
// ./engine \
// -name="Bitflow-Script-Execution-Engine-0" \
// -script="input -> avg() -> output" \
// -input="" \
// -output="bitflow/engine/0" \
// -broker="tcp://192.168.178.20:1883" \
func main() {
	flag.StringVar(&engine.Config.Name, "name", "Unknown script-execution-engine...", "name of engine")
	flag.StringVar(&engine.Config.Script, "script", "input -> output", "script to run")
	flag.StringVar(&engine.Config.InputTopic, "input", "foo", "input MQTT topic")
	flag.StringVar(&engine.Config.OutputTopic, "output", "foo", "output MQTT topic")
	flag.StringVar(&engine.Config.MqttBroker, "broker", "foo", "mqtt broker authority")
	flag.Parse()

	fmt.Println(engine.Config)
}