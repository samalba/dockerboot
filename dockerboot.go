package main

import (
	"log"
)

func main() {
	config := readConfig()
	log.Printf("Config: %#v", config)
	services := parseYmlFile(config.FigFile)
	//TODO(samalba): monitor the file for hot-reload
	if err := runNewServicesState(config, &services); err != nil {
		log.Printf("Cannot process the changes: %s", err)
	}
}
