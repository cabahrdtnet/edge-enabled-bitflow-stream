// +build package

package engine

// NOTE: this test requires docker and the eclipse mosquitto clients and killall

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	// timing variable for potential manual test observation
	startTime  = time.Second * 3
	timeBetweenEvents = time.Millisecond * 100
	eventCount = 25

	// connection
	brokerPort = "18833"
	brokerIP = "127.0.0.1"

	// test data
	incorrectProcessedEvents = []string{}
	numberOfValueDescriptorRequests = 0
	reference = `{"device":"engine-0","origin":000000000000,"readings":[{"origin":000000000000,"name":"humancount","value":"142542"},{"origin":000000000000,"name":"humancount_avg","value":"142542"},{"origin":000000000000,"name":"caninecount","value":"142542"},{"origin":000000000000,"name":"caninecount_avg","value":"142542"}]}`

	//docker     = []string{"docker", "run", "--rm", "-i", "-t",
	//	"cburki/mosquitto-clients" +
	//	"@sha256:89781e1d95c045a273c0be44ef1a6907afa7c3298e0756e0697af6a3de1ecf3f"}
)

func TestRun_AverageCase_SuccessfulRun(t *testing.T) {
	// arrange
	// see if docker is installed, error if not
	// start broker
	defer func() {
		if r := recover(); r != nil {
			tearDownBroker()
			t.Errorf("[engine-test] Recovered in test because of error: %s", r)
			t.Errorf("[engine-test] Please restart test.")
		}
	}()
	arrange()

	// act
	Config.EngineName = "engine-0"
	Config.Script = `input -> append_latency() -> avg() -> output`
	Config.InputTopic = "bitflow/engine/0/source"
	Config.OutputTopic = "bitflow/engine/0/sink"
	Config.CommandTopic = "bitflow/engine/0/command"
	Config.ReverseCommandTopic = "bitflow/engine/0/reverse-command"
	Config.ReverseCommandResponseTopic = "bitflow/engine/0/reverse-command-response"
	Config.MqttBroker = "tcp://" + brokerIP + ":" + brokerPort
	Config.Parameters = "-v -buf 20000"

	Configure()
	exitCode := Run()

	// clean up and assert
	cleanup()
	if exitCode != 0 || len(incorrectProcessedEvents) != 0 {
		t.Errorf("Exit code does not equal 0, but is equal %d", exitCode)
	}

	for _, event := range incorrectProcessedEvents {
		t.Errorf("Wrong processed event!" +
			"\n-Result was: %s" +
			"\n-Expected was: %s",
			event, reference)
	}

	// output header: time,tags,humancount,humancount_avg,caninecount,caninecount_avg
	// 4 registrations and 4 cleaning requests
	//if numberOfValueDescriptorRequests != 8 {
	//	t.Errorf("Wrong number of cleaned up value descriptors!" +
	//		"\n-Result was: %d" +
	//		"\n-Expected was: %d",
	//		numberOfValueDescriptorRequests, 8)
	//}
}

func arrange() {
	setupBroker()
	go publishEvents()
	go subscribeReverseCommand()
	go subscribeProcessedEvents()
}

func cleanup() {
	tearDownBroker()
	killAll()
}

func killAll() {
	killAll := []string{"killall","-SIGINT","mosquitto_sub"}
	cmd := exec.Command(killAll[0], killAll[1:]...)
	if err := cmd.Run(); err != nil {
		fmt.Printf("[engine-test] %s\n", err)
	}
}

func tearDownBroker() {
	docker := []string{"docker", "stop", "engine-test-broker"}
	cmd := exec.Command(docker[0], docker[1:]...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

// docker run -d --rm --name engine-test-broker -p 18833:1883 eclipse-mosquitto
func setupBroker() {
	broker := []string{"docker", "run", "-d", "--rm",
		"--name", "engine-test-broker", "-p", fmt.Sprintf("%s:1883", brokerPort),
		"eclipse-mosquitto"}
	cmd := exec.Command(broker[0], broker[1:]...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

// run dockerized
// 'mosquitto_pub -p 18833
// -i engine-test-data-publisher
// -t 'engine-test/countcamera1/humancount'
// -m {"device":"countcamera1","origin":1471806386919,
// 		"readings":[{"name":"humancount","value":"142542","origin":1471806386919},
// 		{"name":"caninecount","value":"234523450","origin":1471806386919}]}
func publishEvents() {
	fmt.Printf("[engine-test] start sending events in %s\n", startTime)
	timer := time.NewTimer(startTime)
	<-timer.C
	fmt.Printf("[engine-test] start sending events every %s\n", timeBetweenEvents)
	for i := 0; i < eventCount; i++ {
		timer := time.NewTimer(timeBetweenEvents)
		<-timer.C
		fmt.Printf("[engine-test] sending event number: %d\n", i)
		topic := "bitflow/engine/0/source"
		subscriber := "engine-test-data-publisher"
		msg := `{"device":"countcamera1","origin":1471806386919, "readings":[{"name":"humancount","value":"142542","origin":1471806386919},{"name":"caninecount","value":"234523450","origin":1471806386919}]}`
		mosquittoPub := []string{"mosquitto_pub",
			"-p", brokerPort, "-i", subscriber, "-t", topic, "-m", msg}
		cmd := exec.Command(mosquittoPub[0], mosquittoPub[1:]...)
		fmt.Printf("[engine-test] sending event %s\n", msg)
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}
	sendShutdownCommand()
}

// run dockerized
// mosquitto_pub -p 18833
// -i engine-test-command-publisher
// -t 'bitflow/engine/0/command'
// -m 'shutdown'
func sendShutdownCommand() {
	fmt.Printf("[engine-test] sending shutdown command.")
	topic := "bitflow/engine/0/command"
	subscriber := "engine-test-command-publisher"
	msg := "shutdown"
	mosquittoPub := []string{"mosquitto_pub",
		"-p", brokerPort, "-i", subscriber, "-t", topic, "-m", msg}
	cmd := exec.Command(mosquittoPub[0], mosquittoPub[1:]...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

// run dockerized
// mosquitto_sub -v -p 18833
// -i engine-test-reverse-command-subscriber
// -t 'bitflow/engine/0/reverse-command'
func subscribeReverseCommand() {
	topic := "bitflow/engine/0/reverse-command"
	subscriber := "engine-test-reverse-command-subscriber"
	mosquittoSub := []string{"mosquitto_sub",
		"-p", brokerPort, "-i", subscriber, "-t", topic}

	cmd := exec.Command(mosquittoSub[0], mosquittoSub[1:]...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("[engine-test] Could not attach to stdout pipe of: %s\n", subscriber))
	}
	revCmd := make(chan string)
	go reader(stdoutPipe, revCmd)
	go publishResponse(revCmd)
	if err := cmd.Run(); err != nil {
		fmt.Printf("[engine-test] Shutting down reverse command subscription because of %s\n", err)
	}
}

func subscribeProcessedEvents() {
	topic := "bitflow/engine/0/sink"
	subscriber := "engine-test-processed-event-subscriber"
	mosquittoSub := []string{"mosquitto_sub",
		"-p", brokerPort, "-i", subscriber, "-t", topic}

	cmd := exec.Command(mosquittoSub[0], mosquittoSub[1:]...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("[engine-test] Could not attach to stdout pipe of: %s\n", subscriber))
	}
	revCmd := make(chan string)
	go reader(stdoutPipe, revCmd)
	go receiveEvent(revCmd)
	if err := cmd.Run(); err != nil {
		fmt.Printf("[engine-test] Shutting down processed event subscription because of %s\n", err)
	}
}

func receiveEvent(data <-chan string) {
	event := <- data

	// event is equal to reference if all timestamps are ignored
	r, err := regexp.Compile(`[0-9]{13}`)
	if err != nil {
		fmt.Printf("[engine-test] %s", err)
	}
	event = r.ReplaceAllString(event, "000000000000")
	if event != reference {
		incorrectProcessedEvents = append(incorrectProcessedEvents, event)
	}
}

// run dockerized
// 'mosquitto_pub -p 18833
// -i engine-test-reverse-command-response-publisher
// -t 'bitflow/engine/0/reverse-command-response' -m "$i"'
func publishResponse(requests <-chan string) {
	var i int64 = 0
	for req := range requests {
		if strings.Contains(req, "register_value_descriptor") {
			fmt.Printf("[engine-test] sending response to request: %s\n", req)
			topic := "bitflow/engine/0/reverse-command-response"
			subscriber := "engine-test-reverse-command-response-publisher"
			msg := strconv.FormatInt(i*1000, 10) + "abc"
			i++
			mosquittoSub := []string{"mosquitto_pub",
				"-p", brokerPort, "-i", subscriber, "-t", topic, "-m", msg}
			cmd := exec.Command(mosquittoSub[0], mosquittoSub[1:]...)
			if err := cmd.Run(); err != nil {
				panic(err)
			}
			numberOfValueDescriptorRequests++
			continue
		}

		if strings.Contains(req, "clean_value_descriptor") {
			fmt.Printf("[engine-test] got clean_value_descriptor request.\n")
			numberOfValueDescriptorRequests++
			continue
		}
	}
}

// read from `pipe` and write to channel `out`
func reader(pipe io.ReadCloser, out chan<- string) {
	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')
	line = line[:len(line)-1]
	out <- line

	for err == nil {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		line = line[:len(line)-1]
		fmt.Printf("[engine-test] received line from stdout of a cmd: %s\n", line)
		out <- line
	}
}