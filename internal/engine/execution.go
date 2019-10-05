package engine

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Run sets up input for bitflow-pipeline, runs it and monitors its output
func Run() int {
	script, err := mapIO(Config.Script)
	if err != nil {
		panic(err)
	}
	params := strings.Split(Config.Parameters, " ")
	args := append(params, script)

	cmd := exec.Command("bitflow-pipeline", args...)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	go stdinWriter(stdinPipe)
	go stdoutReader(stdoutPipe)
	go stderrReader(stderrPipe)

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	return cmd.ProcessState.ExitCode()
}

func stdinWriter(stdinPipe io.WriteCloser) {
	// TODO use value descriptors in event to derive input header
	//b := []byte(header + "\n")
	writer := bufio.NewWriter(stdinPipe)
	//writer.Write(b)
	//writer.Flush()
	//
	for event := range events.incoming {
		fmt.Println("INPUT: ", event)
		sample, err := etos(event)
		if err != nil {
			fmt.Println("Can't convert event to Bitflow CSV data format sample.")
		}
		writer.Write([]byte(sample + "\n"))
		writer.Flush()
	}
	fmt.Println("Closing command stdin as events.incoming has been closed.")
	stdinPipe.Close()
}

func stdoutReader(stdoutPipe io.ReadCloser) {
	reader := bufio.NewReader(stdoutPipe)
	line := ""
	line, err := reader.ReadString('\n')
	header := line
	fmt.Println("OUTPUT HEADER: ", header)
	for err == nil {
		line, err = reader.ReadString('\n')
		fmt.Println("OUTPUT: ", line)
		// TODO fix: this channel should contain strings in EdgeX event json format
		// convert from Bitflow CSV data format to EdgeX event object
		// TODO use output header to derive value descriptors
		if err != nil {
			close(events.outgoing)
			break
		}
		line = line[:len(line)-1]
		event, convErr := stoe(Config.Name, line, header)
		if convErr != nil {
			panic(convErr)
		}
		// TODO stop exec here once wait for VD uploading
		fmt.Println("This event would be published: ", event)
		events.outgoing <- event
	}
	fmt.Println("Command stdout has been closed because of:", err)
}

func stderrReader(stderrPipe io.ReadCloser) {
	reader := bufio.NewReader(stderrPipe)
	line, err := reader.ReadString('\n')
	for err == nil {
		fmt.Print(line)
		line, err = reader.ReadString('\n')
	}
	fmt.Println("Command stderr has been closed because of:", err)
}