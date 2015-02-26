package main

import (
	"flag"
	"log"
	"os"
	"reflect"
	"strings"
)

type Config struct {
	FigFile   string `default:"./fig.yml"`
	DockerUrl string `default:"unix:///var/run/docker.sock"`
}

func configDefault(fieldName string) string {
	c := Config{}
	st := reflect.TypeOf(c)
	field, found := st.FieldByName(fieldName)
	if !found {
		log.Printf("Invalid config variable `%s'", fieldName)
		return ""
	}
	return field.Tag.Get("default")
}

func readConfigEnvVar(key string, configVar *string) {
	val := os.Getenv(strings.ToUpper(key))
	if val == "" {
		*configVar = configDefault(key)
		return
	}
	*configVar = val
}

func readConfigFromEnv(config *Config) {
	readConfigEnvVar("FigFile", &config.FigFile)
	readConfigEnvVar("DockerUrl", &config.DockerUrl)
}

func readConfigFromFlags(config *Config) {
	flag.StringVar(&config.FigFile, "f", configDefault("FigFile"), "Fig yaml file to read the services config from")
	flag.StringVar(&config.DockerUrl, "H", configDefault("DockerUrl"), "URL of the docker daemon")
	flag.Parse()
}

func readConfig() *Config {
	c := &Config{}
	readConfigFromEnv(c)
	readConfigFromFlags(c)
	return c
}
