package main

import (
	"io/ioutil"

	chatops "github.com/mkobaly/slackchatops"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	SlackToken   string
	SlackChannel string
	Actions      []chatops.Action
}

type Condition interface {
	OkToRun() bool
}

// type Action struct {
// 	Command    string
// 	OutputFile string
// }

// Write will save the configuration to the given path
func (c *Config) Write(path string) error {
	bytes, err := yaml.Marshal(c)
	if err == nil {
		return ioutil.WriteFile(path, bytes, 0777)
	}
	return err
}

// Print will dump the configuration to a string
func (c *Config) Print() (string, error) {
	bytes, err := yaml.Marshal(c)
	if err == nil {
		return string(bytes), nil
	}
	return "", err
}

//NewConfig creates a new Configuration object needed
func NewConfig() *Config {
	//config := Config{}
	var config = &Config{
		SlackToken: "<YOUR SLACK BOT TOKEN>",
	}
	return config
}

//LoadConfig will load up a Config object based on configPath
func LoadConfig(configPath string) *Config {
	//config := Config{}
	var config = new(Config)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err.Error())
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err.Error())
	}
	return config
}
