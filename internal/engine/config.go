package engine

import (
	"fmt"
	"regexp"
	"strings"
)

// these values are set in cmd/device-bitflow/engine/main.go
type Configuration struct {
	Name string
	Script string
	InputTopic string
	OutputTopic string
	MqttBroker string
	Arguments string
}

// replace input and output word in Configuration string by std://-
func ReplaceIO(script string) (string, error) {
	if script == "" {
		return script, fmt.Errorf("script is empty")
	}
	const std = "std://-"
	scriptCopy := script

	unchanged := scriptCopy
	r, _ := regexp.Compile(`\s+`)
	scriptCopy = r.ReplaceAllString(scriptCopy, " ")
	scriptCopy = strings.TrimSpace(script)

	unchanged = scriptCopy
	r, _ = regexp.Compile(`^input`)
	scriptCopy = r.ReplaceAllString(scriptCopy, std)
	if scriptCopy == unchanged {
		return script, fmt.Errorf("leading input designator not found")
	}

	unchanged = scriptCopy
	r, _ = regexp.Compile(`output$`)
	scriptCopy = r.ReplaceAllString(scriptCopy, std)
	if scriptCopy == unchanged {
		return script, fmt.Errorf("trailing output designator not found")
	}

	return scriptCopy, nil
}