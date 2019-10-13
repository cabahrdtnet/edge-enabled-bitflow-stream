package objects

import (
	"context"
	"fmt"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

type Rule struct {
	Name string         `json:"name"`
	Condition Condition `json:"condition"`
	Action Action       `json:"action"`
	Log string          `json:"log"`
}

type Condition struct {
	Device string      `json:"device"`
	Checks []Parameter `json:"checks"`
}

type Parameter struct {
	Value string     `json:"parameter"`
	Operand1 string  `json:"operand1"`
	Operation string `json:"operation"`
	Operand2 string  `json:"operand2"`
}

type Action struct {
	Device string  `json:"device"`
	Command string `json:"command"`
	Body string    `json:"body"`
}

// add rule to rules engine
func (r *Rule) Add() error {
	url := config.URL.RulesEngine + clients.ApiBase + "/rule"
	_, err := clients.PostJsonRequest(url, r, context.TODO())
	if err != nil {
		return fmt.Errorf("couldn't send rule to rules engine: %v", err)
	}
	return nil
}

// remove rule from rules engine
func (r *Rule) Remove() error {
	url := config.URL.RulesEngine + clients.ApiBase + "/rule/name/" + r.Name
	err := clients.DeleteRequest(url, context.TODO())
	if err != nil {
		return fmt.Errorf("couldn't send rule to rules engine: %v", err)
	}
	return nil
}