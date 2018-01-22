package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/zyedidia/micro/cmd/micro/shellwords"
)

// ExecCommand executes a command using exec
// It returns any output/errors
func ExecCommand(name string, arg ...string) (string, error) {
	var err error
	cmd := exec.Command(name, arg...)
	outputBytes := &bytes.Buffer{}
	cmd.Stdout = outputBytes
	cmd.Stderr = outputBytes
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait() // wait for command to finish
	outstring := outputBytes.String()
	return outstring, err
}

// RunShellCommand executes a shell command and returns the output/error
func RunShellCommand(input string) (string, error) {
	args, err := shellwords.Split(input)
	if err != nil {
		return "", err
	}
	inputCmd := args[0]

	return ExecCommand(inputCmd, args[1:]...)
}

func RunBackgroundShell(input string) {
	args, err := shellwords.Split(input)
	if err != nil {
		messenger.Error(err)
		return
	}
	inputCmd := args[0]
	messenger.Message("Running...")
	go func() {
		output, err := RunShellCommand(input)
		totalLines := strings.Split(output, "\n")

		if len(totalLines) < 3 {
			if err == nil {
				messenger.Message(inputCmd, " exited without error")
			} else {
				messenger.Message(inputCmd, " exited with error: ", err, ": ", output)
			}
		} else {
			messenger.Message(output)
		}
		// We have to make sure to redraw
		RedrawAll()
	}()
}

func RunInteractiveShell(input string, wait bool, getOutput bool) (string, error) {
	args, err := shellwords.Split(input)
	if err != nil {
		return "", err
	}
	inputCmd := args[0]

	// Shut down the screen because we're going to interact directly with the shell
	screen.Fini()
	screen = nil

	args = args[1:]

	// Set up everything for the command
	outputBytes := &bytes.Buffer{}
	cmd := exec.Command(inputCmd, args...)
	cmd.Stdin = os.Stdin
	if getOutput {
		cmd.Stdout = io.MultiWriter(os.Stdout, outputBytes)
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	// This is a trap for Ctrl-C so that it doesn't kill micro
	// Instead we trap Ctrl-C to kill the program we're running
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cmd.Process.Kill()
		}
	}()

	cmd.Start()
	err = cmd.Wait()

	output := outputBytes.String()

	if wait {
		// This is just so we don't return right away and let the user press enter to return
		TermMessage("")
	}

	// Start the screen back up
	InitScreen()

	return output, err
}

// HandleShellCommand runs the shell command
// The openTerm argument specifies whether a terminal should be opened (for viewing output
// or interacting with stdin)
func HandleShellCommand(input string, openTerm bool, waitToFinish bool) string {
	if !openTerm {
		RunBackgroundShell(input)
		return ""
	} else {
		output, _ := RunInteractiveShell(input, waitToFinish, false)
		return output
	}
}
