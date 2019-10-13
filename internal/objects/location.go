package objects

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type location struct {
	EngineName string
	Target string
}

// decide where offloading occurs, either locally or remotely
func (l *location) infer(condition string) error {
	fileName := l.EngineName + "-" + "offload.go"
	defer os.Remove(fileName)
	data := []byte(condition)
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("couldn't write temp offload condition file %s: %v", fileName, err)
	}

	goBin := "go"
	args := []string{"run", fileName}
	cmd := exec.Command(goBin, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("couldn't write temp offload condition file %s: %v", fileName, err)
	}
	output := string(outputBytes)
	l.Target = output
	return nil
}