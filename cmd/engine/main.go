package main

import (
	"flag"
	"fmt"
	"github.com/datenente/device-bitflow/internal/engine"
	"os"
)

/*
- script
- input endpoint
- output endpoint
- mqtt broker authority
- engine name
 */
// ./engine \
//// -name="Bitflow-Script-Execution-Engine-0" \
//// -script="input -> avg() -> append-latency() -> output" \
//// -input="" \
//// -output="bitflow/engine/0" \
//// -command="bitflow/engine/0/command" \
//// -broker="tcp://192.168.178.20:1883" \
//// -bitflow-args="-v -buf 20000"

func main() {
	// TODO add possibility to read from toml file
	// TODO improve robustness of process? e.g. bitflow breaks when header isn't the very first thing to arrive
	// parse configuration from CLI arguments
	flag.StringVar(&engine.Config.Name, "name", "Unknown script-execution-engine...", "name of engine")
	flag.StringVar(&engine.Config.Script, "script", "input -> output", "script to run")
	flag.StringVar(&engine.Config.InputTopic, "input", "", "input MQTT topic")
	flag.StringVar(&engine.Config.OutputTopic, "output", "", "output MQTT topic")
	flag.StringVar(&engine.Config.CommandTopic, "command", "", "command MQTT topic")
	flag.StringVar(&engine.Config.MqttBroker, "broker", "", "mqtt broker authority")
	flag.StringVar(&engine.Config.Parameters, "bitflow-args", "", "arguments for bitflow-pipeline process")
	flag.Parse()
	fmt.Println(engine.Config)

	// configure run
	engine.Configure()

	// run
	exitCode := engine.Run()
	os.Exit(exitCode)
}