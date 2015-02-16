package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/samalba/dockerclient"
	"github.com/flynn/go-shlex"
)

func loadCurrentServicesState(dc *dockerclient.DockerClient, c *Config) (Services, error) {
	services := make(Services)
	containers, err := dc.ListContainers(true, false, "")
	if err != nil {
		log.Printf("Cannot list containers: %s", err)
		return nil, err
	}
	containerNames := []string{}
	for _, container := range containers {
		containerInfo, err := dc.InspectContainer(container.Id)
		if err != nil {
			log.Printf("Cannot inspect container id %s, ignoring...", container.Id)
			continue
		}
		service := Service{id: container.Id}
		service.Name = container.Names[0][1:]
		service.Image = containerInfo.Config.Image
		service.Command = container.Command
		containerNames = append(containerNames, service.Name)
		for _, bind := range containerInfo.HostConfig.Binds {
			service.Volumes = append(service.Volumes, bind)
		}
		for containerPort, portBindings := range containerInfo.HostConfig.PortBindings {
			for _, portBinding := range portBindings {
				//NOTE(samalba): we don't want "tcp" in the port name (only udp is explicit)
				containerPort = strings.Replace(containerPort, "/tcp", "", 1)
				port := fmt.Sprintf("%s:%s", portBinding.HostPort, containerPort)
				if portBinding.HostIp != "" {
					port = fmt.Sprintf("%s:%s", portBinding.HostIp, port)
				}
				service.Ports = append(service.Ports, port)
			}
		}
		services[service.Name] = service
	}
	log.Printf("Discovered possible existing services: %s", strings.Join(containerNames, ", "))
	return services, nil
}

func createService(service Service, dc *dockerclient.DockerClient) {
	// Parse command
	cmd, err := shlex.Split(newService.)
	containerConfig := &dockerclient.ContainerConfig{
		Cmd: cmd
	}
	hostConfig := &dockerclient.HostConfig{}
	containerId, err := dc.CreateContainer(containerConfig, service.Name)
	if err != nil {
		return err
	}
	newService.id = containerId
	if err = dc.StartContainer(containerId, hostConfig); err != nil {
		return err
	}
	return err
}

func runNewServicesState(config *Config, newServices *Services) error {
	dc, err := dockerclient.NewDockerClient(c.DockerUrl, nil)
	if err != nil {
		return fmt.Errorf("Cannot connect to the docker daemon (%s): %s", c.DockerUrl, err)
	}
	currentServices, err := loadCurrentServicesState(dc, config)
	if err != nil {
		return err
	}
	log.Println(currentServices)
	for newServiceName, newService := range *newServices {
		currentService, exists := currentServices[newServiceName]
		if exists {
			newService.id = currentService.id
			if reflect.DeepEqual(currentService, newService) {
				// Existing service is equal to the new one, do nothing
				log.Printf("Service `%s' does not change", newServiceName)
				continue
			}
			// New service is different, assuming a conf update, removing...
			log.Printf("Removing existing service `%s' (update)...", newServiceName)
			dc.StopContainer(currentService.id, 10)
			dc.RemoveContainer(currentService.id, true)
		}
		// Create new Service
		if err := createService(newService, dc); err != nil {
			log.Printf("Cannot create Service `%s': %s", newServiceName, err)
			continue
		}
	}
	//Look for changes between newServices and currentServices
	//Fetch what needs to be removed, what needs to be added, what needs to be modified
	return nil
}
