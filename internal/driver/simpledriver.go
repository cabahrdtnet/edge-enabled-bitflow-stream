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
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type SimpleDriver struct {
	lc           logger.LoggingClient
	asyncCh      chan<- *dsModels.AsyncValues
	switchButton bool
}

func getImageBytes(imgFile string, buf *bytes.Buffer) error {
	// Read existing image from file
	img, err := os.Open(imgFile)
	if err != nil {
		return err
	}
	defer img.Close()

	// TODO: Attach MediaType property, determine if decoding
	//  early is required (to optimize edge processing)

	// Expect "png" or "jpeg" image type
	imageData, imageType, err := image.Decode(img)
	if err != nil {
		return err
	}
	// Finished with file. Reset file pointer
	img.Seek(0, 0)
	if imageType == "jpeg" {
		err = jpeg.Encode(buf, imageData, nil)
		if err != nil {
			return err
		}
	} else if imageType == "png" {
		err = png.Encode(buf, imageData)
		if err != nil {
			return err
		}
	}
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *SimpleDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	s.lc = lc
	s.asyncCh = asyncCh
	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *SimpleDriver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", "hello you...")

	if len(reqs) != 1 {
		err = fmt.Errorf("SimpleDriver.HandleReadCommands; too many command requests; only one supported")
		return
	}
	s.lc.Debug(fmt.Sprintf("SimpleDriver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes))
	fmt.Fprintf(os.Stderr, "error: %s\n", "hello you...")

	res = make([]*dsModels.CommandValue, 1)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if reqs[0].DeviceResourceName == "SwitchButton" {
		cv, _ := dsModels.NewBoolValue(reqs[0].DeviceResourceName, now, s.switchButton)
		res[0] = cv
	} else if reqs[0].DeviceResourceName == "Image" {
		// Show a binary/image representation of the switch's on/off value
		buf := new(bytes.Buffer)
		if s.switchButton == true {
			err = getImageBytes("./res/on.png", buf)
		} else {
			err = getImageBytes("./res/off.jpg", buf)
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
func (s *SimpleDriver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	drName := reqs[0].DeviceResourceName
	cmdMessage := "Got command: " + drName + " with params: "
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
	//	err := fmt.Errorf("SimpleDriver.HandleWriteCommands; too many command requests; only one supported")
	//	return err
	//}
	//if len(params) != 1 {
	//	err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the number of parameter is not correct; only one supported")
	//	return err
	//}
	//
	//s.lc.Debug(fmt.Sprintf("SimpleDriver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[0].DeviceResourceName, params))
	//var err error
	//if s.switchButton, err = params[0].BoolValue(); err != nil {
	//	err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be Boolean, parameter: %s", params[0].String())
	//	return err
	//}
	//
	//return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *SimpleDriver) Stop(force bool) error {
	s.lc.Debug(fmt.Sprintf("SimpleDriver.Stop called: force=%v", force))
	return nil
}
