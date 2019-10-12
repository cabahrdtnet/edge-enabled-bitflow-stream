// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// CHANGED BY CHRISTIAN ALEXANDER BAHRDT
// This file is derivative of
// https://github.com/edgexfoundry/device-mqtt-go/blob/edinburgh/internal/driver/config.go
package config

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	Protocol = "mqtt"
)

var (
	URL = struct {
		CoreData string
		CoreMetadata string
		CoreCommand string
		ExportClient string
		RulesEngine string
	}{}
)

type ConnectionInfo struct {
	Schema   string
	Host     string
	Port     string
	User     string
	Password string
	ClientId string
	Topic    string
}

type Configuration struct {
	BrokerSchema    string
	BrokerHost      string
	BrokerPort      string
	CoreDataSchema  string
	CoreDataHost    string
	CoreDataPort    string
	CoreMetadataSchema  string
	CoreMetadataHost    string
	CoreMetadataPort    string
	CoreCommandSchema string
	CoreCommandHost string
	CoreCommandPort string
	ExportClientSchema string
	ExportClientHost string
	ExportClientPort string
	RulesEngineSchema string
	RulesEngineHost string
	RulesEnginePort string
}

// CreateDriverConfig use to load driver config for incoming listener and response listener
func CreateDriverConfig(configMap map[string]string) (*Configuration, error) {
	config := new(Configuration)
	err := load(configMap, config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// CreateConnectionInfo use to load MQTT connectionInfo for read and write command
func CreateConnectionInfo(protocols map[string]models.ProtocolProperties) (*ConnectionInfo, error) {
	info := new(ConnectionInfo)
	protocol, ok := protocols[Protocol]
	if !ok {
		return info, fmt.Errorf("unable to load config, '%s' not exist", Protocol)
	}

	err := load(protocol, info)
	if err != nil {
		return info, err
	}
	return info, nil
}

// load by reflect to check map key and then fetch the value
func load(config map[string]string, des interface{}) error {
	errorMessage := "unable to load config, '%s' not exist"
	val := reflect.ValueOf(des).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		valueField := val.Field(i)

		val, ok := config[typeField.Name]
		if !ok {
			return fmt.Errorf(errorMessage, typeField.Name)
		}

		switch valueField.Kind() {
		case reflect.Int:
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			valueField.SetInt(int64(intVal))
		case reflect.String:
			valueField.SetString(val)
		default:
			return fmt.Errorf("none supported value type %v ,%v", valueField.Kind(), typeField.Name)
		}
	}
	return nil
}

