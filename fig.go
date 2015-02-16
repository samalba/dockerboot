package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// A Service is a fig-defined container
type Service struct {
	id      string
	Name    string
	Image   string
	Command string
	Ports   []string
	Volumes []string
}

type Services map[string]Service

func parseYmlFile(filename string) Services {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Cannot read file `%s': %s", filename, err)
	}
	services := make(Services)
	if err = yaml.Unmarshal(data, &services); err != nil {
		log.Fatalf("Cannot read yaml file `%s': %s", filename, err)
	}
	for name, service := range services {
		service.Name = name
		services[name] = service
	}
	return services
}
