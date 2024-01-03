package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"flag"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
func ParseFlags(rc *ResolverConfiguration) error {
	viper.SetConfigName("syringe")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	flag.UintVar(&rc.TimeoutMillisecons, "TimeoutMillisecons", 5000, "The timeout in milliseconds after which the request should be considered as missing")
	flag.UintVar(&rc.RetryTimes, "RetryTimes", 0, "Amount of retried after a failed request")
	flag.StringVar(&rc.ResolverIp, "ResolverIp", "127.0.0.1", "The resolver to which requests should be sent")
	flag.UintVar(&rc.PinMinTtl, "PinMinTtl", 5, "If the returned ttl is greater than this value, use this ttl instead of the returned value")
	flag.UintVar(&rc.StaticDelaySeconds, "StaticDelaySeconds", 600, "Dont send requests for the configured amount of seconds if the request returns a permanent error")
	flag.UintVar(&rc.FlexibleDelayMinTtlSeconds, "FlexibleDelayMinTtlSeconds", 120, "If a flexible ttl is requested, return a value >= this value")
	flag.UintVar(&rc.FlexibleDelayMaxTtlSeconds, "FlexibleDelayMaxTtlSeconds", 300, "If a flexible ttl is requested, return a value <= this value")
	flag.Uint64Var(&rc.SleepLowTresholdMilliseconds, "SleepLowTresholdMilliseconds", 500, "If there is a bigger gap (>= value) between now and the next due request, sleep for value ms")
	flag.Uint64Var(&rc.SleepLowTresholdCheckIntervalMilliseconds, "SleepLowTresholdCheckIntervalMilliseconds", 50, "If in next due time is < SleepLowTresholdMilliseconds, poll the queue every value ms")
	flag.UintVar(&rc.ServerListenPort, "ServerListenPort", 9000, "Port on which the webserver should listen")
	flag.StringVar(&rc.DomainsFile, "DomainsFile", "domains.txt", "A file which contains the list of domains to preheat. Entries must be separated by newline '\\n'. Syntax 'domain rrtype' (e.g. 'github.com A')")
	flag.BoolVar(&rc.LoadDomainsFileOnStart, "LoadDomainsFileOnStart", true, "Load the domains file on start")
	flag.UintVar(&rc.LoadDomainsFileInitialQueryLimit, "LoadDomainsFileInitialQueryLimit", 500, "Limit to value requests per second when reading from DomainsFile")
	flag.UintVar(&rc.LogLevel, "LogLevel", 3, "LogLevel (1-8) to use. 1=Panic,8=Trace - see https://github.com/sirupsen/logrus")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return viper.Unmarshal(rc)
}

type ResolverConfiguration struct {
	TimeoutMillisecons                        uint   `yaml:"TimeoutMillisecons"`
	RetryTimes                                uint   `yaml:"RetryTimes"`
	ResolverIp                                string `yaml:"ResolverIp"`
	PinMinTtl                                 uint   `yaml:"PinMinTtl"`
	StaticDelaySeconds                        uint   `yaml:"StaticDelaySeconds"`
	FlexibleDelayMinTtlSeconds                uint   `yaml:"FlexibleDelayMinTtlSeconds"`
	FlexibleDelayMaxTtlSeconds                uint   `yaml:"FlexibleDelayMaxTtlSeconds"`
	SleepLowTresholdMilliseconds              uint64 `yaml:"SleepLowTresholdMilliseconds"`
	SleepLowTresholdCheckIntervalMilliseconds uint64 `yaml:"SleepLowTresholdCheckIntervalMilliseconds"`
	ServerListenPort                          uint   `yaml:"ServerListenPort"`
	DomainsFile                               string `yaml:"DomainsFile"`
	LoadDomainsFileOnStart                    bool   `yaml:"LoadDomainsFileOnStart"`
	LoadDomainsFileInitialQueryLimit          uint   `yaml:"LoadDomainsFileInitialQueryLimit"`
	LogLevel                                  uint   `yaml:"LogLevel"`
}

func StructToKeyValuePairs(config *ResolverConfiguration) map[string]interface{} {
	pairs := make(map[string]interface{})
	v := reflect.ValueOf(*config)
	for i := 0; i < v.NumField(); i++ {
		pairs[v.Type().Field(i).Name] = v.Field(i).Interface()
	}
	return pairs
}
