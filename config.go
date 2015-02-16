package main

import (
	"os"
	"strings"
)

type Config struct {
	Figfile             string
	DockerUrl           string
}

func readConfigEnvVar(key string, configVar *string, defaultValue string) {
	val := os.Getenv(strings.ToUpper(key))
	if val == "" {
		*configVar = defaultValue
		return
	}
	*configVar = val
}

func loadConfigFromEnv() *Config {
	c := &Config{}
	readConfigEnvVar("figfile", &c.Figfile, "./fig.yml")
	readConfigEnvVar("dockerurl", &c.DockerUrl, "unix:///var/run/docker.sock")
	return c
}
