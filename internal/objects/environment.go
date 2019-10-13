package objects

import (
	"fmt"
	"github.com/datenente/device-bitflow/internal/config"
	"github.com/datenente/device-bitflow/internal/naming"
	"os"
)

type environment struct {
	Target string
}

func (e environment) set() error {
	switch e.Target {
	case "local":
		set(naming.DockerTLSVerify, config.Docker.LocalDockerTLSVerify)
		set(naming.DockerHost, config.Docker.LocalDockerHost)
		set(naming.DockerCertPath, config.Docker.LocalDockerCertPath)
		set(naming.DockerMachineName, config.Docker.LocalDockerMachineName)
		return nil

	case "remote":
		set(naming.DockerTLSVerify, config.Docker.RemoteDockerTLSVerify)
		set(naming.DockerHost, config.Docker.RemoteDockerHost)
		set(naming.DockerCertPath, config.Docker.RemoteDockerCertPath)
		set(naming.DockerMachineName, config.Docker.RemoteDockerMachineName)
		return nil

	default:
		return fmt.Errorf("unknown target for setting environment: %s", e.Target)
	}
}

func (e environment) unset() {
	unset(naming.DockerTLSVerify)
	unset(naming.DockerHost)
	unset(naming.DockerCertPath)
	unset(naming.DockerMachineName)
}

func (e *environment) get(key string) (string, error) {
	string := os.Getenv(key)
	if string == "" {
		return "", fmt.Errorf("env variable %s is empty", key)
	}
	return string, nil
}

func set(key string, env string) error {
	err := os.Setenv(key, env)
	if err != nil {
		return fmt.Errorf("couldn't set %s env variable: %v", env, err)
	}
	return nil
}

func unset(key string) error {
	err := os.Unsetenv(key)
	if err != nil {
		return fmt.Errorf("couldn't unset %s env variable: %v", key, err)
	}
	return nil
}
