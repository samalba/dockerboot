package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/samalba/dockerclient"
)

type CliCmd struct {
	Help string
	Func func(*Config, *dockerclient.DockerClient, Services, Services) error
}

var commands = map[string]CliCmd{
	"start":   {"Start all services and process configuration updates", cmdStart},
	"reload":  {"Process configuration updates (alias to `start')", cmdStart},
	"stop":    {"Stop all services", cmdStop},
	"restart": {"Restart all services and process configuration updates", cmdRestart},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] COMMAND\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	for name, cmd := range commands {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", name, cmd.Help)
	}
	os.Exit(0)
}

func cmdStart(config *Config,
	dc *dockerclient.DockerClient,
	currentServices Services,
	newServices Services) error {
	if err := startServices(config, dc, currentServices, newServices); err != nil {
		log.Printf("Cannot start the services: %s", err)
		return err
	}
	return nil
}

func cmdStop(config *Config,
	dc *dockerclient.DockerClient,
	currentServices Services,
	newServices Services) error {
	if err := stopServices(config, dc, currentServices, newServices); err != nil {
		log.Printf("Cannot stop the services: %s", err)
		return err
	}
	return nil
}

func cmdRestart(config *Config,
	dc *dockerclient.DockerClient,
	currentServices Services,
	newServices Services) error {
	if err := cmdStop(config, dc, currentServices, newServices); err != nil {
		return err
	}
	if err := cmdStart(config, dc, currentServices, newServices); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Usage = usage
	config := readConfig()
	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}
	cmd, exists := commands[strings.ToLower(args[0])]
	if !exists {
		flag.Usage()
	}
	log.Printf("%#v", *config)
	newServices := parseYmlFile(config.FigFile)
	dc, err := initDockerClient(config)
	if err != nil {
		log.Fatalf("Cannot connect to the docker daemon (%s): %s", config.DockerUrl, err)
	}
	currentServices, err := loadCurrentServicesState(config, dc)
	if err != nil {
		log.Fatalf("Cannot load existing services: %s", err)
	}
	cmd.Func(config, dc, currentServices, newServices)
}
