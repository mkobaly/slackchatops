package main

import (
	"io/ioutil"
	"runtime"

	chatops "github.com/mkobaly/slackchatops"
	yaml "gopkg.in/yaml.v2"
)

// Config represents all of the settings needed to run the chatOps application
type Config struct {
	SlackToken   string
	SlackChannel string
	Actions      []chatops.Action
}

//TODO: Not used yet. Ideally want to have conditions for Actions. Say approval needed before running action
// type Condition interface {
// 	OkToRun() bool
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
	var config = &Config{
		SlackToken:   "<YOUR SLACK BOT TOKEN>",
		SlackChannel: "<SLACK CHANNEL>",
	}
	var actions []chatops.Action
	if runtime.GOOS == "windows" {
		actions = append(actions, chatops.Action{Name: "ip", Command: "cmd", Description: "Show your IP address(s)", Args: []string{"/C", "ipconfig"}})
		actions = append(actions, chatops.Action{Name: "dir", Command: "cmd", Description: "List directory content", AuthorizedUsers: []string{"UC1111111"}, Args: []string{"/C", "dir"}})
	} else {
		actions = append(actions, chatops.Action{Name: "ip", Command: "ifconfig", Description: "Show your IP address(s)"})
		actions = append(actions, chatops.Action{Name: "ls", Command: "ls", Description: "List directory content", AuthorizedUsers: []string{"UC1111111"}, Args: []string{"-la"}})
	}
	config.Actions = actions
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
