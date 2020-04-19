package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/zyedidia/micro/internal/screen"
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

// RunCommand executes a shell command and returns the output/error
func RunCommand(input string) (string, error) {
	args, err := shellquote.Split(input)
	if err != nil {
		return "", err
	}
	if len(args) == 0 {
		return "", errors.New("No arguments")
	}
	inputCmd := args[0]

	return ExecCommand(inputCmd, args[1:]...)
}

// RunBackgroundShell runs a shell command in the background
// It returns a function which will run the command and returns a string
// message result
func RunBackgroundShell(input string) (func() string, error) {
	args, err := shellquote.Split(input)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, errors.New("No arguments")
	}
	inputCmd := args[0]
	return func() string {
		output, err := RunCommand(input)
		totalLines := strings.Split(output, "\n")

		str := output
		if len(totalLines) < 3 {
			if err == nil {
				str = fmt.Sprint(inputCmd, " exited without error")
			} else {
				str = fmt.Sprint(inputCmd, " exited with error: ", err, ": ", output)
			}
		}
		return str
	}, nil
}

// executeCmdPipe executes a stack of commands
// This function executes a stack of chained commands using io.Pipes.
//
// TODO
// Implement other type of command support?
// egs.
//	If first ok: cmd1 && cmd2
//  Sequential:  cmd1  & cmd2
//  Concurrent:  cmd1  ; cmd2
func executeCmdPipe(outputBuffer *bytes.Buffer, getOutput bool, stack ...*exec.Cmd) (err error) {
	if len(stack) == 0 {
		return fmt.Errorf("length of the cmd stack is zero (0)")
	}

	var errorBuffer bytes.Buffer
	var pipeLen int
	var pipes []*io.PipeWriter

	// Handle the simple case, i.e. no pipes
	if len(stack) == 1 {
		pipeLen = 1
		stack[0].Stdin = os.Stdin

		if getOutput {
			stack[0].Stdout = io.MultiWriter(os.Stdout, outputBuffer)
		} else {
			stack[0].Stdout = os.Stdout
		}

		stack[0].Stderr = os.Stderr
		stack[0].Start()

		log.Println("ExecuteCmdPipe: stack ->", stack)
		log.Println("ExecuteCmdPipe: pipes ->", pipes)
		log.Println("ExecuteCmdPipe: pipeLen     ->", pipeLen)
		log.Println("ExecuteCmdPipe: want output ->", getOutput)

		return stack[0].Wait()
	}

	pipeLen = len(stack) - 1
	pipes = make([]*io.PipeWriter, pipeLen)

	i := 0

	for ; i < pipeLen; i++ {
		stdinPipe, stdoutPipe := io.Pipe()
		stack[i].Stdout = stdoutPipe
		stack[i].Stderr = &errorBuffer
		stack[i+1].Stdin = stdinPipe
		pipes[i] = stdoutPipe
	}

	if getOutput {
		stack[i].Stdout = outputBuffer
	} else {
		stack[i].Stdout = os.Stdout
	}

	stack[i].Stderr = os.Stderr

	log.Println("ExecuteCmdPipe: stack ->", stack)
	log.Println("ExecuteCmdPipe: pipes ->", pipes)
	log.Println("ExecuteCmdPipe: pipeLen     ->", pipeLen)
	log.Println("ExecuteCmdPipe: want output ->", getOutput)

	if err := executePipe(stack, pipes); err != nil {
		return fmt.Errorf("executePipe(%s): %v", errorBuffer.Bytes(), err)
	}

	return err

}

// executePipe is a helper recursive function for ExecuteCmdPipe
func executePipe(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}

	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}

		defer func() {
			if err == nil {
				pipes[0].Close()
				err = executePipe(stack[1:], pipes[1:])
			}
		}()
	}

	return stack[0].Wait()
}

// RunInteractiveShell runs a shellcommand interactively
func RunInteractiveShell(input string, wait bool, getOutput bool) (string, error) {
	args, err := shellquote.Split(input)
	if err != nil {
		return "", err
	}

	if len(args) == 0 {
		return "", errors.New("No arguments")
	}

	commands := strings.Split(input, "|")

	// Shut down the screen because we're going to interact
	// directly with the shell and defer the screen backup
	screenb := screen.TempFini()
	defer screen.TempStart(screenb)

	outputBuffer := &bytes.Buffer{}

	stack := make([]*exec.Cmd, len(commands))

	for i := range commands {
		iArgs, err := shellquote.Split(commands[i])
		if err != nil {
			return "", err
		}

		stack[i] = exec.Command(iArgs[0], iArgs[1:]...)
	}

	log.Println("RunInteractiveShell: stack ->", stack)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			for i := range args {
				if stack[i].Process != nil {
					stack[i].Process.Kill()
				}
			}
		}
	}()

	err = executeCmdPipe(outputBuffer, getOutput, stack...)
	if err != nil {
		log.Println("RunInteractiveShell: got error ->", err)
	}

	log.Println("RunInteractiveShell: outputBuffer ->", outputBuffer)

	if wait {
		// This is just so we don't return right away and let the user press enter to return
		screen.TermMessage("")
	}

	return outputBuffer.String(), err
}

// UserCommand runs the shell command
// The openTerm argument specifies whether a terminal should be opened (for viewing output
// or interacting with stdin)
// func UserCommand(input string, openTerm bool, waitToFinish bool) string {
// 	if !openTerm {
// 		RunBackgroundShell(input)
// 		return ""
// 	} else {
// 		output, _ := RunInteractiveShell(input, waitToFinish, false)
// 		return output
// 	}
// }
