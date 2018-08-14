package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nlopes/slack"
	logrus "github.com/sirupsen/logrus"

	"github.com/fatih/color"
	cmdline "github.com/galdor/go-cmdline"
	chatops "github.com/mkobaly/slackchatops"
	"github.com/shomali11/slacker"
)

const (
	empty   = ""
	space   = " "
	dash    = "-"
	newLine = "\n"
)

var running bool

func main() {
	//Define command line params and parse input
	cmdline := cmdline.New()
	cmdline.AddOption("c", "config", "config.yaml", "Path to configuration file")
	cmdline.Parse(os.Args)

	//Load up configuration. This holds TeamCity and Emitter info
	cfgPath := "./config.yaml"
	if cmdline.IsOptionSet("c") {
		cfgPath = cmdline.OptionValue("c")
	}

	//no config file so create one
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		config := NewConfig()
		config.Write(cfgPath)
		color.Yellow("---------------------------------------------------------------------------------")
		color.Yellow("config.yaml not present. One was just created for you. Please edit it accordingly")
		color.Yellow("---------------------------------------------------------------------------------")
		os.Exit(0)
	}

	log := chatops.NewLogger("chatops")

	// Load up configuration file
	config := LoadConfig(cfgPath)
	bot := slacker.NewClient(config.SlackToken)
	bot.Help(helpHandler(bot, config.SlackChannel))

	for _, a := range config.Actions {
		description := a.Description
		if description == "" {
			description = a.Name
		}
		params := ""
		for _, p := range a.Params {
			params += " <" + p + ">"
		}
		bot.Command(a.Name+params, description, handler(a, config.SlackChannel, log))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if config.SlackChannel != "" {
		color.Yellow("only listening on slack channel " + config.SlackChannel)
	}
	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// overridding default help handler to ensure we only resond to correct channel
func helpHandler(s *slacker.Slacker, channel string) func(slacker.Request, slacker.ResponseWriter) {
	return func(request slacker.Request, response slacker.ResponseWriter) {
		//ensure only running for specified channel
		if channel != "" && channel != request.Event().Channel {
			return
		}
		helpMessage := empty
		for _, command := range s.BotCommands() {
			tokens := command.Tokenize()
			for _, token := range tokens {
				if token.IsParameter {
					helpMessage += fmt.Sprintf("`%s`", token.Word) + space
				} else {
					helpMessage += fmt.Sprintf("`%s`", token.Word) + space
				}
			}
			helpMessage += dash + space + fmt.Sprintf("_%s_", command.Description()) + newLine
		}
		response.Reply(helpMessage)
	}
}

func handler(a chatops.Action, channel string, log *logrus.Entry) func(slacker.Request, slacker.ResponseWriter) {
	return func(request slacker.Request, response slacker.ResponseWriter) {
		//ensure only running for specified channel
		if channel != "" && channel != request.Event().Channel {
			return
		}

		if running {
			response.Reply("Busy with another action. Please wait...")
			return
		}

		log.WithFields(logrus.Fields{"request": request}).Info(a.Name)
		var args []string
		for _, p := range a.Params {
			arg := request.StringParam(p, "")
			parts := strings.Split(arg, " ")
			args = append(args, parts...)
		}

		running = true
		response.Typing()
		result, err := a.Run(args...)
		running = false

		response.Reply("*ExitCode: " + strconv.Itoa(result.ReturnCode) + "*")
		response.Reply("_Output:_\n" + result.StdOut)
		if err != nil {
			response.Reply("_Error:_\n" + err.Error())
		}

		//is there a file to upload (say test results)
		if _, err := os.Stat(a.OutputFile); err == nil {
			rtm := response.RTM()
			client := response.Client()

			rtm.SendMessage(rtm.NewOutgoingMessage("Uploading output file ...", request.Event().Channel))
			client.UploadFile(slack.FileUploadParameters{File: a.OutputFile, Channels: []string{channel}})
			os.Remove(a.OutputFile)
		}
	}
}
