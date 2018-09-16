package slackchatops

import (
	"bytes"
	"log"
	"os/exec"
	"syscall"
)

// Action represents what the system should perform. This is typically some type of command
type Action struct {
	Name            string   // friendly name of the action
	Description     string   // description of the action
	Command         string   // actual command being called
	WorkingDir      string   // working directory for the command to be called in
	Params          []string // parameters the command needs to run. When executed the user will pass these in as arguments. They will be appended to the Args list
	Args            []string // arguments to pass to the command. If any are predefined in the config.yaml file (defaults) then user passed arguments (Params) will be appended to the end
	OutputFile      string   // if the command being executed writes to a file. StdErr and StdOut are already captured. This could be an html document from a set of unit tests for example
	AuthorizedUsers []string // list of autorized users that are allowed to execute this action. This should be their slackId
}

// Result of an Action being executed on the system
type Result struct {
	ReturnCode int
	StdOut     string
	StdError   string
}

// Run actually executes the command
func (a *Action) Run(args ...string) (Result, error) {
	cmd := exec.Command(a.Command, append(a.Args, args...)...)
	if a.WorkingDir != "" {
		cmd.Dir = a.WorkingDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	outStr, errStr := stdout.String(), stderr.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			log.Printf("Could not get exit code for failed program: %v, %v", a.Command, a.Args)
			exitCode = 1
			if errStr == "" {
				errStr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return Result{
		ReturnCode: exitCode,
		StdError:   errStr,
		StdOut:     outStr,
	}, err
}
