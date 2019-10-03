package engine

import (
	"bufio"
	"os/exec"
	//"io"
	"io/ioutil"
	//"bytes"
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type BitflowCommand struct {

}

// TODO find strategy how to set this path
// PIPELINEPATH = path to bitflow-pipeline
// ENGINEPATH = path to engine

func Run() {
	Config.Script, _ = ReplaceIO(Config.Script)
	//TODO reset script
	//script := Config.Script
	//script := `std://- -> avg() -> tcp+csv://:55555`
	// This isn't directly documented anywhere, but it works.
	script := `std://- -> avg() -> append_latency() -> avg() -> std+csv://-`
	//script := `std://- -> std+csv://-`

	// create script file
	d1 := []byte(script+"\n")
	err := ioutil.WriteFile("/tmp/dat1", d1, 0644)
	check(err)

	// setup environment
	os.Setenv("BITFLOWPATH", "/Users/pinie/go/src/github.com/bitflow-stream/go-bitflow/cmd/bitflow-pipeline")
	os.Setenv("PATH", os.Getenv("PATH")+":"+os.Getenv("BITFLOWPATH"))
	//fmt.Println("PATH:", os.Getenv("PATH"))

	// command
	cmd := exec.Command("bitflow-pipeline", "-f", "/tmp/dat1", "-qq")
	//cmd := exec.Command("grep", "--line-buffered", "o")
	//cmd := exec.Command("nc", "-l", "55555", "-k")

	//s2 := "2017-11-09 13:51:09.877210495,experiment=cpu host=wally133,1,2,3,4,4"
	//s3 := "2017-11-09 13:51:09.877210495,experiment=cpu host=wally133,1,2,3,4,4"
	//s := [3]string{s1, s2, s3}

	ipipe, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	// TODO add error checks
	opipe, _ := cmd.StdoutPipe()

	// TODO add error checks
	epipe, _ := cmd.StderrPipe()
	//cmd.Run
	if err := cmd.Start(); err != nil {
		fmt.Println(err.Error())
	}

	go func () {
		// TODO add check if channel is closed
		msgh := <- data.Subscription
		fmt.Println("INPUT HEADER: ", msgh)
		b := []byte(msgh+"\n")
		ipipe.Write(b)
		for msg := range data.Subscription {
			fmt.Println("INPUT: ", msg)
			b := []byte(msg+"\n")
			ipipe.Write(b)
			if err != nil {
				fmt.Print("error")
			}
		}
		// this signals bitflow that it should stop reading input
		// effectively closing it gracefully
		fmt.Println("Closing ipipe.")
		ipipe.Close()
		fmt.Println(cmd.ProcessState.ExitCode())
		fmt.Println("Finished.")
		//header := "time,tags,cpu,disk-io/all/io,disk-io/all/ioBytes,disk-io/all/ioTime,disk-io/all/sausage\n"
		//s1 := "2017-11-09 13:51:09.877210495,experiment=cpu host=wally133,0,0,0,0,0\n"
		//ipipe.Write([]byte(header))
		//for {
		//	time.Sleep(1 * time.Second)
		//	b := []byte(s1)
		//	ipipe.Write(b)
		//	//ipipe.Close()
		//	if err != nil {
		//		return
		//		fmt.Print("error")
		//	}
		//}
	}()

	go func () {
		//for {
		//	b := []byte{}
		//	opipe.Read(b)
		//	if len(b) == 0 {
		//		//fmt.Println("Is empty...")
		//	} else {
		//		data.Publication <- string(b)
		//	}
		//}

		reader := bufio.NewReader(opipe)
		line := ""
		//line, err := reader.ReadString('\n')
		// for err == nil
		line, err := reader.ReadString('\n')
		fmt.Print("OUTPUT HEADER: ", line)
		i := 0
		for err == nil {
			fmt.Println("READING")
			line, err = reader.ReadString('\n')
			if err != nil {
				close(data.Publication)
				break
			}
			fmt.Print("OUTPUT: ", line)
			i++
			fmt.Println("[[", i, "]]")
			line = line[:len(line)-1]
			//publish(line)
			//fmt.Println("Waiting for line.")
			data.Publication <- line
		}

		// TODO remove line this has already happened
		//opipe.Close()
		// TODO remove next line this is redundant
		if err != nil {
			fmt.Println("Command output pipe has been closed. Reason:", err)
		}
	}()

	// TODO put this in a go routine
	// read stderr of bitflow (instead of stdout as this is potentially used by executing scripts)
	reader := bufio.NewReader(epipe)
	line, err := reader.ReadString('\n')
	for err == nil {
		fmt.Print("L ", line)
		line, err = reader.ReadString('\n')
	}

	if err != nil {
		fmt.Println("Command error pipe has been closed. Reason:", err)
	}

	fmt.Println("Closing script execution engine.")
}