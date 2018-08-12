# SlackChatOps

Easy to use Slack bot used for chatOps. This bot can be configured to run any number of actions via Slack
and report the results back. All actions are configured via the config.yaml file

## Building

Clone repo and run
```
./build.sh
```
This will generate windows and linux binaries in the bin folder

Run the chatops executable

## Configuration

Upon first run the application will exit and inform you the configuration file (config.yaml) was not present. 
A sample config file will be created for you in the current directory. By default a few actions have been created
as samples

An action is defined below. 
```
// Action represents what the system should perform. This is typically some type of command
type Action struct {
	Name        string   // friendly name of the action
	Description string   // description of the action
	Command     string   // actual command being called
	WorkingDir  string   // working directory for the command to be called in
	Params      []string // parameters the command needs to run. When executed the user will pass these in as arguments. They will be appended to the Args list
	Args        []string // arguments to pass to the command. If any are predefined in the config.yaml file (defaults) then user passed arguments (Params) will be appended to the end
	OutputFile  string   // if the command being executed writes to a file. StdErr and StdOut are already captured. This could be an html document from a set of unit tests for example
}
```

## Slack Setup

Within your slack application click the  "+ Add Apps" link and browse for  'Bots'. That URL should be

https://[YOUR_DOMAIN].slack.com/apps/A0F7YS25R-bots

Click "Add Configuration" and choose a username for your Bot (Assuming chatops)

Copy the API Token as you will need to add that to your configuration created above

### Channel specific bot

Setting the "slackchannel" value is optional and will make this bot only respond to commands for the given channel. THIS IS RECOMMENDED or else if you run multiple chatBots they all will respond.


## Typical setup

Suppose you create private channels for Development & Production (chatOps-dev & chatOps-prod)

* Identify the slack channel by looking at the URLs. For example in the below url the channel is GC6AAAAAA
```
https://yourdomain.slack.com/messages/GC6AAAAAA/team/U0000000/
```

* Run this chatops application and configure the Slack API token and Slack Channel
* Create custom tasks for your development stack as needed. For example
    - Deploying new versions of your application
    - Running integration tests
    - Reading log files

    Note: You have the freedom to create any task you need for your environment. If you can run it via a shell script or command line arguments
    it can be run here and made available though Slack

* Via Slack type the below command to list out available actions the chatBot can perform
```
@chatops help
```

## TODOs

- [ ] Permission restricted actions. Useful for production actions
- [ ] Custom output formatters for Slack
