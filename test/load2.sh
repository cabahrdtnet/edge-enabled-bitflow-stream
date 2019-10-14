#! /usr/bin/env fish

http PUT http://localhost:48082/api/v1/device/name/engine-2/command/script contents="input -> append_latency() -> avg() -> output"
http PUT http://localhost:48082/api/v1/device/name/engine-2/command/source devices="countcamera1,countcamera2" value_descriptors="humancount"                                                                                                    
http PUT http://localhost:48082/api/v1/device/name/engine-2/command/offloading condition='package main ; import "fmt" ; func main() { fmt.Print("local") }'
http PUT http://localhost:48082/api/v1/device/name/engine-2/command/actuation actuation_device_name="engine-2" command_name="script" command_body='{"action" : "stop"}' actuation_left_operand="Integer.parseInt(value)" actuation_operator=">" actuation_right_operand="5"
http PUT http://localhost:48082/api/v1/device/name/engine-2/command/control action="start"
