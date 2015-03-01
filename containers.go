package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/flynn/go-shlex"
	"github.com/samalba/dockerclient"
)

func initDockerClient(config *Config) (*dockerclient.DockerClient, error) {
	dc, err := dockerclient.NewDockerClient(config.DockerUrl, nil)
	if err != nil {
		return nil, err
	}
	return dc, nil
}

func loadCurrentServicesState(c *Config, dc *dockerclient.DockerClient) (Services, error) {
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
		service.Volumes = containerInfo.HostConfig.Binds
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

func parsePorts(portsStr []string) (map[string][]dockerclient.PortBinding, error) {
	ports := make(map[string][]dockerclient.PortBinding)
	for _, port := range portsStr {
		p := strings.Split(port, ":")
		lnP := len(p)
		if lnP < 2 || lnP > 3 {
			return nil, fmt.Errorf("Invalid port format: %s", port)
		}
		portBinding := dockerclient.PortBinding{}
		if lnP == 3 {
			portBinding.HostIp = p[0]
		}
		portBinding.HostPort = p[lnP-2]
		portName := p[lnP-1]
		if strings.Contains(portName, "/") {
			pN := strings.SplitN(portName, "/", 2)
			if pN[1] != "udp" {
				return nil, fmt.Errorf("Container port should specify the protocol only if udp: %s", port)
			}
		} else {
			portName = portName + "/tcp"
		}
		if _, exists := ports[portName]; exists {
			ports[portName] = []dockerclient.PortBinding{}
		}
		ports[portName] = append(ports[portName], portBinding)
	}
	return ports, nil
}

func createService(dc *dockerclient.DockerClient, service *Service) error {
	// Parse command
	cmd, err := shlex.Split(service.Command)
	if err != nil {
		return err
	}
	containerConfig := &dockerclient.ContainerConfig{
		Tty:       true,
		OpenStdin: true,
		Cmd:       cmd,
		Image:     service.Image,
	}
	log.Printf("Creating service `%s'...", service.Name)
	containerId, err := dc.CreateContainer(containerConfig, service.Name)
	if err != nil {
		return err
	}
	service.id = containerId
	return nil
}

func startService(dc *dockerclient.DockerClient, service *Service) error {
	ports, err := parsePorts(service.Ports)
	if err != nil {
		return err
	}
	hostConfig := &dockerclient.HostConfig{
		Binds:        service.Volumes,
		PortBindings: ports,
	}
	log.Printf("Starting service `%s'...", service.Name)
	if err = dc.StartContainer(service.id, hostConfig); err != nil {
		return err
	}
	return nil
}

func startServices(config *Config,
	dc *dockerclient.DockerClient,
	currentServices Services,
	newServices Services) error {
	for newServiceName, newService := range newServices {
		currentService, exists := currentServices[newServiceName]
		if exists {
			newService.id = currentService.id
			if reflect.DeepEqual(currentService, newService) {
				// Existing service is equal to the new one, do nothing...
				log.Printf("Service `%s' does not change", newServiceName)
				// ...other than starting it
				if err := startService(dc, &newService); err != nil {
					log.Printf("Cannot start Service `%s': %s", newServiceName, err)
					continue
				}
				continue
			}
			// New service is different, assuming a conf update, removing...
			log.Printf("Removing existing service `%s' (update)...", newServiceName)
			dc.StopContainer(currentService.id, 10)
			dc.RemoveContainer(currentService.id, true)
		}
		// Create new Service
		if err := createService(dc, &newService); err != nil {
			log.Printf("Cannot create Service `%s': %s", newServiceName, err)
			continue
		}
		// Start the service
		if err := startService(dc, &newService); err != nil {
			log.Printf("Cannot start Service `%s': %s", newServiceName, err)
			continue
		}
	}
	return nil
}

func stopServices(config *Config,
	dc *dockerclient.DockerClient,
	currentServices Services,
	newServices Services) error {
	for newServiceName, _ := range newServices {
		currentService, exists := currentServices[newServiceName]
		if exists {
			log.Printf("Stopping service `%s'...", newServiceName)
			dc.StopContainer(currentService.id, 5)
		}
	}
	return nil
}
