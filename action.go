package slackchatops

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
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
	mergedArgs := a.ParseArgs(args)
	cmd := exec.Command(a.Command, mergedArgs...)
	if a.WorkingDir != "" {
		path, _ := ExpandPath(a.WorkingDir)
		cmd.Dir = path
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

// ValidateArgs will ensure all tokenized parameters {x} have been replaced
func (a *Action) ValidateArgs() error {
	//ensure given number of params we have same number of tokens
	for i := 0; i < len(a.Params); i++ {
		p := "{" + strconv.Itoa(i) + "}"
		valid := false
		for _, ar := range a.Args {
			if strings.Contains(ar, p) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("Action %s is missing argument %s for parameter %s", a.Name, p, a.Params[i])
		}
	}

	//Now parse args and ensure
	args := a.ParseArgs(a.Params)
	for i := 0; i < 20; i++ { //choosing arbitrary number (20)
		p := "{" + strconv.Itoa(i) + "}"
		for _, ar := range args {
			if strings.Contains(ar, p) {
				return fmt.Errorf("Action %s has too many tokenized arguments. %s is not used", a.Name, p)
			}
		}
	}
	return nil
}

// ParseArgs will combine the the user input with the parameters to
// generate the final argument list
func (a *Action) ParseArgs(args []string) []string {
	result := []string{}
	result = append(result, a.Args...)
	if len(args) == 0 {
		return result
	}

	for i, arg := range args {
		for j, argDef := range result {
			replace := "{" + strconv.Itoa(i) + "}"
			result[j] = strings.Replace(argDef, replace, arg, -1)
		}
	}
	return result
}

// func (a *Action) MergeArgs(args []string) ([]string, error) {
// 	var result []string
// 	if len(args) == 0 {
// 		return a.Args, nil
// 	}
// 	for _, arg := range a.Args {
// 		if strings.HasPrefix(arg, "$") {
// 			arg1 := strings.Replace(arg, "$", "", -1)
// 			i, err := strconv.Atoi(arg1)
// 			if err != nil {
// 				return nil, err
// 			}
// 			result = append(result, args[i-1])
// 		} else {
// 			result = append(result, arg)
// 		}
// 	}
// 	return result, nil
// }

func ExpandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}
