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
	header := <-data.subscription

	b := []byte(header + "\n")
	writer := bufio.NewWriter(stdinPipe)
	writer.Write(b)
	writer.Flush()
	for msg := range data.subscription {
		// TODO fix: this channel should contain strings in Bitflow CSV data format
		// msg = event?
		// convert here via etos
		//actmsg := etos(msg)
		fmt.Println("INPUT: ", msg)
		b := []byte(msg + "\n")
		writer.Write(b)
		writer.Flush()
	}
	fmt.Println("Closing command stdin as data.subscription has been closed.")
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
		event, convErr := stoe(Config.Name, line, header)
		if convErr != nil {
			panic(convErr)
		}
		fmt.Println("This event would be published: ", event)

		if err != nil {
			close(data.publication)
			break
		}
		line = line[:len(line)-1]
		data.publication <- line
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