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
	"github.com/datenente/device-bitflow-stream/internal/communication"
	conf "github.com/datenente/device-bitflow-stream/internal/config"
	"github.com/datenente/device-bitflow-stream/internal/naming"
	"github.com/datenente/device-bitflow-stream/internal/objects"
	"github.com/edgexfoundry/device-sdk-go"
	"os"
	"strconv"
	"strings"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	driver *Driver
)

type Driver struct {
	lc      logger.LoggingClient
	asyncCh chan<- *dsModels.AsyncValues
	config  *conf.Configuration
	registry objects.Registry
	// see [@GopherCon2017Lightning]
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	s.lc = lc
	s.asyncCh = asyncCh
	conf.Log = lc
	driver = s

	s.registry = objects.Registry{}
	err := s.registry.Init()
	if err != nil {
		panic(fmt.Errorf("could not init registry: %v", err))
	}

	config, err := conf.CreateDriverConfig(device.DriverConfigs())
	if err != nil {
		panic(fmt.Errorf("could not read driver configuration: %v", err))
	}
	s.config = config

	// set URLs of other microservices bundled with EdgeX
	conf.URL.CoreData = s.config.CoreDataSchema + "://" + s.config.CoreDataHost + ":" + s.config.CoreDataPort
	conf.URL.CoreMetadata = s.config.CoreMetadataSchema + "://" + s.config.CoreMetadataHost + ":" + s.config.CoreMetadataPort
	conf.URL.CoreCommand = s.config.CoreCommandSchema + "://" + s.config.CoreCommandHost + ":" + s.config.CoreCommandPort
	conf.URL.ExportClient = s.config.ExportClientSchema + "://" + s.config.ExportClientHost + ":" + s.config.ExportClientPort
	conf.URL.RulesEngine = s.config.RulesEngineSchema + "://" + s.config.RulesEngineHost + ":" + s.config.RulesEnginePort

	// set local and remote docker machine
	conf.Docker.LocalDockerTLSVerify = s.config.LocalDockerTLSVerify
	conf.Docker.LocalDockerHost = s.config.LocalDockerHost
	conf.Docker.LocalDockerCertPath = s.config.LocalDockerCertPath
	conf.Docker.LocalDockerMachineName = s.config.LocalDockerMachineName
	conf.Docker.RemoteDockerTLSVerify = s.config.RemoteDockerTLSVerify
	conf.Docker.RemoteDockerHost = s.config.RemoteDockerHost
	conf.Docker.RemoteDockerCertPath = s.config.RemoteDockerCertPath
	conf.Docker.RemoteDockerMachineName = s.config.RemoteDockerMachineName

	// set broker config
	brokerPort64, err := strconv.ParseInt(s.config.BrokerPort, 10, 64)
	if err != nil {
		panic(fmt.Errorf("couldn't parse broker port in because of: %v", err))
	}
	conf.Broker.Name = "MqttBroker"
	conf.Broker.Schema = s.config.BrokerSchema
	conf.Broker.Host = s.config.BrokerHost
	conf.Broker.Port = int(brokerPort64)

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
	index, err := naming.ExtractIndex(deviceName, "-", 1)
	if err != nil {
		err = fmt.Errorf("couldn't index from device name: %v", err)
		return err
	}

	command := "nop"
	if reqs[0].DeviceResourceName == "action" {
		command = "control"
	}

	if reqs[0].DeviceResourceName == "contents" {
		command = "script"
	}

	if reqs[0].DeviceResourceName == "devices" || reqs[0].DeviceResourceName == "value_descriptors" {
		command = "source"
	}

	if reqs[0].DeviceResourceName == "actuation_device_name" || reqs[0].DeviceResourceName == "command_name" ||
		reqs[0].DeviceResourceName == "command_body" || reqs[0].DeviceResourceName == "actuation_left_operand" ||
		reqs[0].DeviceResourceName == "actuation_operator" || reqs[0].DeviceResourceName == "actuation_right_operand" {
		command = "actuation"
	}

	if reqs[0].DeviceResourceName == "condition" {
		command = "offloading"
	}

	switch command {
	case "control":
		action, err := params[0].StringValue()
		if err != nil {
			err = fmt.Errorf("couldn't determine parameter action of %s command: %v", params[0].DeviceResourceName, err)
			return err
		}

		if action == "start" {
			err = s.registry.Start(index)
			if err != nil {
				return fmt.Errorf("couldn't start engine in HandleWriteCommands: %v", err)
			}
			return nil
		}

		if action == "stop" {
			err = s.registry.Stop(index)
			if err != nil {
				return fmt.Errorf("couldn't stop engine in HandleWriteCommands: %v", err)
			}
			return nil
		}

		return fmt.Errorf("couldn't determine action %s of %s command: %v", action, params[0].DeviceResourceName, err)

	case "script":
		contents, err := params[0].StringValue()
		if err != nil {
			err = fmt.Errorf("couldn't determine parameter contents of %s command: %v", params[0].DeviceResourceName, err)
			return err
		}

		template := objects.Engine{
			Configuration: objects.EngineConfiguration{
				Script: contents,
			},
		}

		err = s.registry.Update(index, template)
		if err != nil {
			err = fmt.Errorf("couldn't update %s with template %v", deviceName, template)
			return err
		}

		return nil

	case "source":
		type resource struct {
			name string
			value string
		}

		resources := [2]resource{}
		for i := 0; i < 2; i++ {
			if reqs[i].DeviceResourceName == "devices" {
				devices, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter devices of %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[0].name = reqs[i].DeviceResourceName
				resources[0].value = devices
			}

			if reqs[i].DeviceResourceName == "value_descriptors" {
				valueDescriptors, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter value_descriptors of %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[1].name = reqs[i].DeviceResourceName
				resources[1].value = valueDescriptors
			}
		}

		devices := resources[0].value
		valueDescriptors := resources[1].value

		inputDeviceNames := strings.Split(strings.TrimSpace(devices), ",")
		inputValueDescriptorNames := strings.Split(strings.TrimSpace(valueDescriptors), ",")

		template := objects.Engine{
			Configuration: objects.EngineConfiguration{
				InputDeviceNames:          inputDeviceNames,
				InputValueDescriptorNames: inputValueDescriptorNames,
			},
		}

		err = s.registry.Update(index, template)
		if err != nil {
			err = fmt.Errorf("couldn't update %s with template %v", deviceName, template)
			return err
		}

		return nil

	case "actuation":
		type resource struct {
			name string
			value string
		}

		resources := [6]resource{}
		for i := 0; i < 6; i++ {
			if reqs[i].DeviceResourceName == "actuation_device_name" {
				actuationDeviceName, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter actuation_device_name of %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[0].name = reqs[i].DeviceResourceName
				resources[0].value = actuationDeviceName
			}

			if reqs[i].DeviceResourceName == "command_name" {
				commandName, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter command_name of %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[1].name = reqs[i].DeviceResourceName
				resources[1].value = commandName
			}

			if reqs[i].DeviceResourceName == "command_body" {
				commandBody, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter command_body of %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[2].name = reqs[i].DeviceResourceName
				resources[2].value = commandBody
			}

			if reqs[i].DeviceResourceName == "actuation_left_operand" {
				leftOperand, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter actuation_left_operand %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[3].name = reqs[i].DeviceResourceName
				resources[3].value = leftOperand
			}

			if reqs[i].DeviceResourceName == "actuation_operator" {
				operator, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter actuation_operator %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[4].name = reqs[i].DeviceResourceName
				resources[4].value = operator
			}

			if reqs[i].DeviceResourceName == "actuation_right_operand" {
				rightOperand, err := params[i].StringValue()
				if err != nil {
					err = fmt.Errorf("couldn't determine parameter actuation_right_operand %s command: %v", params[i].DeviceResourceName, err)
					return err
				}
				resources[5].name = reqs[i].DeviceResourceName
				resources[5].value = rightOperand
			}
		}

		actuationDeviceName := resources[0].value
		commandName         := resources[1].value
		commandBody         := resources[2].value
		leftOperand         := resources[3].value
		operator            := resources[4].value
		rightOperand        := resources[5].value

		actuation := objects.Actuation{
			DeviceName:   actuationDeviceName,
			CommandName:  commandName,
			CommandBody:  commandBody,
			LeftOperand:  leftOperand,
			Operator:     operator,
			RightOperand: rightOperand,
		}

		template := objects.Engine{
			Configuration: objects.EngineConfiguration{
				Actuation: actuation,
			},
		}

		err = s.registry.Update(index, template)
		if err != nil {
			err = fmt.Errorf("couldn't update %s with template %v", deviceName, template)
			return err
		}

		return nil

	case "offloading":
		offloadCondition, err := params[0].StringValue()
		if err != nil {
			err = fmt.Errorf("couldn't determine parameter condition of %s command: %v", params[0].DeviceResourceName, err)
			return err
		}

		template := objects.Engine{
			Configuration: objects.EngineConfiguration{
				OffloadCondition: offloadCondition,
			},
		}

		err = s.registry.Update(index, template)
		if err != nil {
			err = fmt.Errorf("couldn't update %s with template %v", deviceName, template)
			return err
		}

		return nil

	case "nop":
		return fmt.Errorf("couldn't recognize any device command other than \"nop\"")
	}

	return fmt.Errorf("couldn't recognize any device command")
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *Driver) Stop(force bool) error {
	s.lc.Debug(fmt.Sprintf("Driver.Stop called: force=%v", force))
	return nil
}
