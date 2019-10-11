// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// CHANGED BY CHRISTIAN ALEXANDER BAHRDT
// this file is derivative of
// https://github.com/edgexfoundry/device-sdk-go/blob/edinburgh/example/driver/simpledriver.go
package driver

import (
	"bytes"
	"fmt"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/models"
	"github.com/edgexfoundry/device-sdk-go"
	"os"
	"strconv"
	"sync"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	driver *Driver
)

type Driver struct {
	lc           logger.LoggingClient
	asyncCh      chan<- *dsModels.AsyncValues
	engines      map[string]models.Engine
	config       *configuration
	mutex        sync.Mutex
	// see [@GopherCon2017Lightning]
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	s.lc = lc
	s.asyncCh = asyncCh
	s.engines = make(map[string]models.Engine)
	driver = s

	err := initEngineRegistry()
	if err != nil {
		panic(fmt.Errorf("could not init registry: %v", err))
	}

	config, err := CreateDriverConfig(device.DriverConfigs())
	if err != nil {
		panic(fmt.Errorf("could not read driver configuration: %v", err))
	}
	s.config = config

	urls.CoreData = s.config.CoreDataSchema + "://" + s.config.CoreDataHost + ":" + s.config.CoreDataPort
	urls.ExportClient = s.config.ExportClientDataSchema + "://" + s.config.ExportClientHost + ":" + s.config.ExportClientPort
	communication.Broker = s.config.BrokerSchema + "://" + s.config.BrokerHost + ":" + s.config.BrokerPort

	go InitRegistrySubscription()
	go handleRegistryRequest()

	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *Driver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", "hello you...")

	if len(reqs) != 1 {
		err = fmt.Errorf("Driver.HandleReadCommands; too many command requests; only one supported")
		return
	}
	s.lc.Debug(fmt.Sprintf("Driver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes))
	fmt.Fprintf(os.Stderr, "error: %s\n", "hello you...")

	res = make([]*dsModels.CommandValue, 1)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if reqs[0].DeviceResourceName == "SwitchButton" {
		// true was s.SwitchButton
		cv, _ := dsModels.NewBoolValue(reqs[0].DeviceResourceName, now, true)
		res[0] = cv
	} else if reqs[0].DeviceResourceName == "Image" {
		// Show a binary/image representation of the switch's on/off value
		buf := new(bytes.Buffer)
		//s.switchButton instead of first true
		if true == true {
			//err = getImageBytes("./res/on.png", buf)
		} else {
			//err = getImageBytes("./res/off.jpg", buf)
		}
		cvb, _ := dsModels.NewBinaryValue(reqs[0].DeviceResourceName, now, buf.Bytes())
		res[0] = cvb
	} else if reqs[0].DeviceResourceName == "output" {
		cvs := dsModels.NewStringValue(reqs[0].DeviceResourceName, now, "this is your results. it's nothing.")
		fmt.Fprintf(os.Stderr, "error: %s\n", "hello you results...")
		s.lc.Info("Got commando message to device resource" +  reqs[0].DeviceResourceName)
		res[0] = cvs
	} else if reqs[0].DeviceResourceName == "status" {
		cvs := dsModels.NewStringValue(reqs[0].DeviceResourceName, now, "current status.")
		fmt.Fprintf(os.Stderr, "error: %s\n", "hello you... status")
		s.lc.Info("Got commando message to device resource" +  reqs[0].DeviceResourceName)
		res[0] = cvs
	} else {
		s.lc.Info("I could not handle anything:" +  reqs[0].DeviceResourceName)
		fmt.Fprintf(os.Stderr, "error: %s\n", "hello you... anything")
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	return
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *Driver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	s.lc.Info(deviceName)
	deviceResourceName := reqs[0].DeviceResourceName

	switch deviceResourceName {
	case "control":
		action, err := params[0].StringValue()
		if err != nil {
			err = fmt.Errorf("couldn't determine action parameter of control command: %v", err)
			return err
		}

		if action == "start" {
			s.lc.Info("Engine is starting! So exciting!")
			return nil
		}

		if action == "stop" {
			s.lc.Info("Engine is stopping! So boring!")
			return nil
		}
	case "source":
		// ["devices","value_descriptors"]
	case "actuation":
		// ["command_name", "body", "actuation_device", "actuation_left_operand", "actuation_operator", "actuation_right_operand"]
	case "offloading":
		// condition
	}

	cmdMessage := "Got command: " + deviceResourceName + " with params: "
	s.lc.Info(cmdMessage)
	for index, param := range params {
		paramMessage := "{"
		paramMessage += "Param No. " + strconv.Itoa(index) + ":"
		paramMessage += param.DeviceResourceName + ":"
		paramMessage += "::"
		stringValue, _ := param.StringValue()
		paramMessage += stringValue
		paramMessage += "}"
		s.lc.Info(paramMessage)
	}

	return nil

	//if len(reqs) != 1 {
	//	err := fmt.Errorf("Driver.HandleWriteCommands; too many command requests; only one supported")
	//	return err
	//}
	//if len(params) != 1 {
	//	err := fmt.Errorf("Driver.HandleWriteCommands; the number of parameter is not correct; only one supported")
	//	return err
	//}
	//
	//s.lc.Debug(fmt.Sprintf("Driver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[0].DeviceResourceName, params))
	//var err error
	//if s.switchButton, err = params[0].BoolValue(); err != nil {
	//	err := fmt.Errorf("Driver.HandleWriteCommands; the data type of parameter should be Boolean, parameter: %s", params[0].String())
	//	return err
	//}
	//
	//return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *Driver) Stop(force bool) error {
	s.lc.Debug(fmt.Sprintf("Driver.Stop called: force=%v", force))
	return nil
}
