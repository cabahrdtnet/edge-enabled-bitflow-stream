package objects

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"strings"
)

type Actuation struct {
	DeviceName string
	CommandName string
	CommandBody string
	LeftOperand string
	Operator string
	RightOperand string
}

// create rule from actuation and device index string to differentiate between identical rules between engines
// there is support for one rule per engine accordingly
func (a *Actuation) inferRule() (Rule, error) {
	// set rule name
	ruleName := a.DeviceName + "-check"

	// set parameter
	parameter := Parameter{
		Value:     a.CommandName,
		Operand1:  a.LeftOperand,
		Operation: a.Operator,
		Operand2:  a.RightOperand,
	}

	// set action
	deviceID, err := a.getIDofDevice()
	if err != nil {
		return Rule{}, fmt.Errorf("couldn't get ID of device with name %s, because of: %v", a.DeviceName, err)
	}
	commandID, err := a.getIDOfCommand()
	if err != nil {
		return Rule{}, fmt.Errorf("couldn't get ID of command with name %s, because of: %v", a.DeviceName, err)
	}

	// rules require a JSON string's quotation marks escaped like this
	body := strings.Replace(a.CommandBody, "\"", "\\\"", -1)
	log := a.DeviceName

	rule := Rule{
		Name: ruleName,
		Condition: Condition{
			Device: a.DeviceName,
			Checks: []Parameter{
				parameter,
			},
		},
		Action:    Action{
			Device:  deviceID,
			Command: commandID,
			Body:    body,
		},
		Log:       log,
	}

	return rule, nil
}

// get ID of device for action in rule
func (a *Actuation) getIDofDevice() (string, error) {
	url := config.URL.CoreMetadata + clients.ApiDeviceRoute + "/name/" + a.DeviceName
	payload, err := clients.GetRequest(url, context.TODO())
	if err != nil {
		return "", fmt.Errorf("couldn't get device %s from core data to get ID: %v", a.DeviceName, err)
	}
	var device models.Device
	err = json.Unmarshal(payload, &device)
	if err != nil {
		return "", fmt.Errorf("couldn't unmarshal device %s to get ID: %v", a.DeviceName, err)
	}
	return device.Id, nil
}

// get ID of command for action in rule
func (a *Actuation) getIDOfCommand() (string, error) {
	ruleName := a.DeviceName + "-check"
	url := config.URL.CoreCommand + clients.ApiDeviceRoute + "/name/" + ruleName
	payload, err := clients.GetRequest(url, context.TODO())
	if err != nil {
		return "", fmt.Errorf("couldn't get command response %s from core command to get command ID: %v", a.DeviceName, err)
	}
	var commandResponse models.CommandResponse
	err = json.Unmarshal(payload, &commandResponse)
	if err != nil {
		return "", fmt.Errorf("couldn't unmarshal command response %s to get command ID: %v", a.DeviceName, err)
	}
	for _, command := range commandResponse.Commands {
		if command.Name == a.CommandName {
			return command.Id, nil
		}
	}
	return "", fmt.Errorf("couldn't find command with name %s for device %s", a.CommandName, a.DeviceName)
}