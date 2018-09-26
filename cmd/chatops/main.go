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
var debugging bool

func main() {
	//Define command line params and parse input
	cmdline := cmdline.New()
	cmdline.AddOption("c", "config", "config.yaml", "Path to configuration file")
	cmdline.AddFlag("d", "debug", "Log additional information for debugging purposes")
	cmdline.Parse(os.Args)

	//Load up configuration
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
	debugging = cmdline.IsOptionSet("d")

	// Load up configuration file
	config := LoadConfig(cfgPath)
	bot := slacker.NewClient(config.SlackToken)
	bot.Help(helpHandler(bot, config.SlackChannel))
	bot.DefaultCommand(func(request slacker.Request, response slacker.ResponseWriter) {

		if config.SlackChannel != "" && config.SlackChannel != request.Event().Channel {
			return
		}
		attachments := []slack.Attachment{}
		attachments = append(attachments, slack.Attachment{
			Color: "warning",
			Title: "Unknown action",
		})
		response.Reply("", slacker.WithAttachments(attachments))
	})

	for _, a := range config.Actions {
		description := a.Description
		if description == "" {
			description = a.Name
		}
		params := ""
		for _, p := range a.Params {
			params += " <" + p + ">"
		}
		err := a.ValidateArgs()
		if err != nil {
			color.Yellow("---------------------------------------------------------------------------------")
			color.Yellow("An action within config.yaml is not parameterized correctly")
			color.Yellow(err.Error())
			color.Yellow("---------------------------------------------------------------------------------")
			os.Exit(1)
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
		debug("In help handler: Channel:" + request.Event().Channel)
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

		debug("In handler: Channel:" + request.Event().Channel)
		//ensure only running for specified channel
		if channel != "" && channel != request.Event().Channel {
			return
		}

		if running {
			response.Reply("Busy with another action. Please wait...")
			return
		}

		authorized := true //Default to true. Only when action has restriction do we limit access

		//This action has authorization check. Ensure user is part of authorized user list
		if len(a.AuthorizedUsers) > 0 {
			authorized = false
			user := request.Event().User
			for _, u := range a.AuthorizedUsers {
				if user == u {
					authorized = true
				}
			}
		}

		if !authorized {
			attachments := []slack.Attachment{}
			attachments = append(attachments, slack.Attachment{
				Color: "danger",
				Title: "You are not authorized to execute this action",
			})
			response.Reply("", slacker.WithAttachments(attachments))
			return
		}

		log.WithFields(logrus.Fields{"command": a.Name}).Info("InHandler")
		var args []string
		for _, p := range a.Params {
			arg := request.StringParam(p, "")
			parts := strings.Split(arg, " ")
			args = append(args, parts...)
		}

		running = true
		response.Typing()
		debugf("Args: %v", args)
		result, err := a.Run(args...)
		running = false

		response.Reply("*ExitCode: " + strconv.Itoa(result.ReturnCode) + "*")
		if result.StdOut != "" {
			response.Reply("_Output:_\n" + result.StdOut)
		}
		if err != nil {
			response.Reply("_Error:_\n" + result.StdError)
		}

		outputFile, _ := chatops.ExpandPath(a.OutputFile)
		//is there a file to upload (say test results)
		if _, err := os.Stat(outputFile); err == nil {
			rtm := response.RTM()
			client := response.Client()

			rtm.SendMessage(rtm.NewOutgoingMessage("Uploading output file ...", request.Event().Channel))
			client.UploadFile(slack.FileUploadParameters{File: outputFile, Channels: []string{channel}})
			os.Remove(outputFile)
		}
	}
}

func debug(msg string) {
	if debugging {
		fmt.Println(msg)
	}
}

func debugf(format string, a ...interface{}) {
	if debugging {
		fmt.Printf(format+"\n", a)
	}
}
