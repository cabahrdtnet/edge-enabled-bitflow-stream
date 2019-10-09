package main

import (
	"flag"
	"fmt"
	"github.com/datenente/device-bitflow/internal/engine"
	"os"
)

func main() {
	// TODO add possibility to read from toml file
	// TODO improve robustness of process? e.g. bitflow breaks when header isn't the very first thing to arrive
	// parse configuration from CLI arguments
	flag.StringVar(&engine.Config.EngineName, "name", "Unknown script-execution-engine...", "name of engine")
	flag.StringVar(&engine.Config.Script, "script", "input -> output", "script to run")
	flag.StringVar(&engine.Config.InputTopic, "input", "", "input MQTT topic")
	flag.StringVar(&engine.Config.OutputTopic, "output", "", "output MQTT topic")
	flag.StringVar(&engine.Config.CommandTopic, "command", "", "command MQTT topic")
	flag.StringVar(&engine.Config.ReverseCommandTopic, "reverse-command", "", "reverse command MQTT topic")
	flag.StringVar(&engine.Config.ReverseCommandResponseTopic, "reverse-command-response", "", "reverse command response MQTT topic")
	flag.StringVar(&engine.Config.MqttBroker, "broker", "", "mqtt broker authority")
	flag.StringVar(&engine.Config.Parameters, "bitflow-params", "", "arguments for bitflow-pipeline process")
	flag.Parse()
	fmt.Println(engine.Config)

	engine.Configure()
	exitCode := engine.Run()
	engine.CleanUp()
	os.Exit(exitCode)
}