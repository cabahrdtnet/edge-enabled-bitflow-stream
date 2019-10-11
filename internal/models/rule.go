package models

import (
	"context"
	"fmt"
	drv "github.com/datenente/device-bitflow/internal/driver"
	"github.com/datenente/device-bitflow/internal/naming"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

//{
//	"name": "randnum_too_huge",
//
//	"condition": {
//	"device": "MQ_DEVICE",
//		"checks": [
//			{
//				"parameter": "randnum",
//				"operand1": "Float.parseFloat(value)",
//				"operation": ">",
//				"operand2": "27.0"
//			}
//		]
//	},
//
//	"action": {
//		"device": "6d2a216d-60b3-42fc-9f3a-37cc7daaab3f",
//		"command": "8e8c92ea-229a-4746-acfd-8a1ebb580cb9",
//		"body": "{\\\"message\\\":\\\"WARNING temp too high!\\\"}"
//	},
//
//	"log": "This random number is definitely too friggin huge."
//}

type Rule struct {
	Name string         `json:"name"`
	Condition Condition `json:"condition"`
	Action Action       `json:"action"`
	Log string          `json:"log"`
}

type Condition struct {
	Checks []Parameter `json:"condition"`
}

type Parameter struct {
	Value string     `json:"parameter"`
	Operand1 string  `json:"operand1"`
	Operation string `json:"operator"`
	Operand2 string  `json:"operand2"`
}

type Action struct {
	Device string  `json:"device"`
	Command string `json:"command"`
	Body string    `json:"body"`
}

// create rule from actuation and device index string to differentiate between identical rules between engines
func ApplyRuleFrom(actuation Actuation, index int64) error {
	name := naming.Name(index)

	parameter := Parameter{
		Value:     actuation.CommandName,
		Operand1:  actuation.LeftOperand,
		Operation: actuation.Operator,
		Operand2:  actuation.RightOperand,
	}

	log := name
	deviceID, err := IDOfDevice(name)
	if err != nil {
		return fmt.Errorf("couldn't get ID of device with index %d, because of: %v", index, err)
	}

	commandID, err := IDOfCommand(name)
	if err != nil {
		return fmt.Errorf("couldn't get ID of command with index %d, because of: %v", index, err)
	}

	rule := Rule{
		Name: name,
		Condition: Condition{
			Checks: []Parameter{
				parameter,
			},
		},
		Action:    Action{
			Device:  deviceID,
			Command: commandID,
			Body:    actuation.CommandBody,
		},
		Log:       log,
	}

	_, err = clients.PostJsonRequest(drv.URL.RulesEngine, rule, context.TODO())
	if err != nil {
		return fmt.Errorf("couldn't send rule to rules engine: %v", err)
	}
	return nil
}

func IDOfDevice(name string) (string, error) {
	payload, err := clients.GetRequest(drv.URL.CoreMetadata, context.TODO())
	ID := string(payload)
	if err != nil {
		return ID, nil
	} else {
		return "", fmt.Errorf("couldn't derive ID for device %s, because of: %v", err)
	}
}

func IDOfCommand(name string) (string, error) {
	payload, err := clients.GetRequest(drv.URL.CoreCommand, context.TODO())
	ID := string(payload)
	if err != nil {
		return ID, nil
	} else {
		return "", fmt.Errorf("couldn't derive ID for command %s, because of: %v", err)
	}
}