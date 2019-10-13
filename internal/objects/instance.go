package objects

import (
	"fmt"
	"github.com/datenente/device-bitflow/internal/communication"
	"github.com/datenente/device-bitflow/internal/naming"
	"os/exec"
)

type Instance struct {
	Engine    Engine
	Environment environment
	location  location
	docker   *exec.Cmd
}

// prepare an instance and execute it
func (i *Instance) Create() error {
	target, err := i.prepare()
	if err != nil {
		return fmt.Errorf("error in instance preparation: %v", err)
	}

	err = i.execute(target)
	if err != nil {
		return fmt.Errorf("error in instance execution: %v", err)
	}
	return nil
}

// prepare execution of instance
func (i *Instance) prepare() (string, error) {
	// set docker environment based on desired execution location
	condition := i.Engine.Configuration.OffloadCondition
	location := location{
		EngineName: i.Engine.Name,
	}
	err := location.infer(condition)
	if err != nil {
		return "", fmt.Errorf("couldn't get execution location: %v", err)
	}

	// docker run config.DockerEngineImage
	dockerArgs := []string{"run", naming.DockerEngineImage}
	instanceArgs := i.args()
	args := append(dockerArgs, instanceArgs...)
	i.docker = exec.Command(naming.DockerCommand, args...)
	return location.Target, nil
}

// execute instance
func (i *Instance) execute(target string) error {
	env := environment{Target:target}
	env.set()
	defer env.unset()

	// TODO err is also not nil if engine shuts down directly after being started without initial event
	if err := i.docker.Start(); err != nil {
		return fmt.Errorf("couldn't run docker instance with args %s: %v", i.args(), err)
	}
	// todo env
	return nil
}

// set arguments of instance, i.e. docker run "instance" args...
func (i *Instance) args() []string {
	name := i.Engine.Name
	script := i.Engine.Configuration.Script
	input := naming.Topic(i.Engine.Index, naming.Source)
	output := naming.Topic(i.Engine.Index, naming.Sink)
	command := naming.Topic(i.Engine.Index, naming.Command)
	reverseCommand := naming.Topic(i.Engine.Index, naming.ReverseCommand)
	reverseCommandResponse := naming.Topic(i.Engine.Index, naming.ReverseCommandResponse)
	broker := communication.Broker

	args := []string{
		fmt.Sprintf("-name=%s", name),
		fmt.Sprintf("-script=%s", script),
		fmt.Sprintf("-input=%s", input),
		fmt.Sprintf("-output=%s", output),
		fmt.Sprintf("-command=%s", command),
		fmt.Sprintf("-reverse-command=%s", reverseCommand),
		fmt.Sprintf("-reverse-command-response=%s", reverseCommandResponse),
		fmt.Sprintf("-broker=%s", broker),
		//-"bitflow-params"="-v -buf 20000"
	}

	return args
}