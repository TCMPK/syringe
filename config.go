package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) *ResolverConfiguration {
	config := &ResolverConfiguration{
		ResolverIp:                                "127.0.0.1",
		TimeoutMillisecons:                        5000,
		RetryTimes:                                0,
		PinMinTtl:                                 10,
		StaticDelaySeconds:                        10,
		FlexibleDelayMinTtlSeconds:                300,
		FlexibleDelayMaxTtlSeconds:                600,
		SleepLowTresholdMilliseconds:              1000,
		SleepLowTresholdCheckIntervalMilliseconds: 50,
		ServerListenPort:                          8000,
		DomainsFile:                               "",
		LoadDomainsFileOnStart:                    false,
		LoadDomainsFileInitialQueryLimit:          100,
		LogLevel:                                  3,
	}

	// Parse the YAML configuration file.
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}

	// Set default values for fields in the Config struct.
	kvPairs := StructToKeyValuePairs(config)
	configString := ""
	for name, hex := range kvPairs {
		configString += fmt.Sprintf("%s: %s ", name, hex)
	}
	log.Warn("Loaded configuration ", strings.Trim(configString, " "))
	return config
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./syringe.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}

type ResolverConfiguration struct {
	TimeoutMillisecons                        uint32 `yaml:"TimeoutMillisecons" default:"5000"`
	RetryTimes                                uint32 `yaml:"RetryTimes" default:"0"`
	ResolverIp                                string `yaml:"ResolverIp" default:"127.0.0.1"`
	PinMinTtl                                 uint32 `yaml:"PinMinTtl" default:"10"`
	StaticDelaySeconds                        uint32 `yaml:"StaticDelaySeconds" default:"600"`
	FlexibleDelayMinTtlSeconds                uint32 `yaml:"FlexibleDelayMinTtlSeconds" default:"120"`
	FlexibleDelayMaxTtlSeconds                uint32 `yaml:"FlexibleDelayMaxTtlSeconds" default:"300"`
	SleepLowTresholdMilliseconds              int64  `yaml:"SleepLowTresholdMilliseconds" default:"500"`
	SleepLowTresholdCheckIntervalMilliseconds int64  `yaml:"SleepLowTresholdCheckIntervalMilliseconds" default:"50"`
	ServerListenPort                          int    `yaml:"ServerListenPort" default:"8000"`
	DomainsFile                               string `yaml:"DomainsFile" default:""`
	LoadDomainsFileOnStart                    bool   `yaml:"LoadDomainsFileOnStart" default:"false"`
	LoadDomainsFileInitialQueryLimit          uint32 `yaml:"LoadDomainsFileInitialQueryLimit" default:"100"`
	LogLevel                                  uint32 `yaml:"LogLevel" default:"3"`
}

func StructToKeyValuePairs(config *ResolverConfiguration) map[string]interface{} {
	pairs := make(map[string]interface{})
	v := reflect.ValueOf(*config)
	for i := 0; i < v.NumField(); i++ {
		pairs[v.Type().Field(i).Name] = v.Field(i).Interface()
	}
	return pairs
}
