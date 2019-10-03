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
// -bitflow-args="-v -buf 20000"

func main() {
	// parse configuration from CLI arguments
	flag.StringVar(&engine.Config.Name, "name", "Unknown script-execution-engine...", "name of engine")
	flag.StringVar(&engine.Config.Script, "script", "input -> output", "script to run")
	flag.StringVar(&engine.Config.InputTopic, "input", "", "input MQTT topic")
	flag.StringVar(&engine.Config.OutputTopic, "output", "", "output MQTT topic")
	flag.StringVar(&engine.Config.MqttBroker, "broker", "", "mqtt broker authority")
	flag.StringVar(&engine.Config.Arguments, "bitflow-args", "", "arguments for bitflow-pipeline process")
	flag.Parse()
	fmt.Println(engine.Config)

	// configure run
	engine.Configure()

	// run
	engine.Run()}