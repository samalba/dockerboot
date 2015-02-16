package main

import (
	"log"
)

func main() {
	config := loadConfigFromEnv()
	log.Printf("Config: %#v", config)
	services := parseYmlFile(config.Figfile)
	//TODO(samalba): monitor the file for hot-reload
	if err := runNewServicesState(config, &services); err != nil {
		log.Printf("Cannot process the changes: %s", err)
	}
}
