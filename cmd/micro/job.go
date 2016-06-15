package main

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
)

type JobFunction struct {
	function func(string, ...string)
	output   string
	args     []string
}

type CallbackFile struct {
	io.Writer

	callback func(string, ...string)
	args     []string
}

func (f *CallbackFile) Write(data []byte) (int, error) {
	jobFunc := JobFunction{f.callback, string(data), f.args}
	jobs <- jobFunc
	return f.Writer.Write(data)
}

func JobStart(cmd string, onStdout, onStderr, onExit string, userargs ...string) *exec.Cmd {
	split := strings.Split(cmd, " ")
	args := split[1:]
	cmdName := split[0]

	proc := exec.Command(cmdName, args...)
	var outbuf bytes.Buffer
	if onStdout != "" {
		proc.Stdout = &CallbackFile{&outbuf, LuaFunctionJob(onStdout), userargs}
	} else {
		proc.Stdout = &outbuf
	}
	if onStderr != "" {
		proc.Stderr = &CallbackFile{&outbuf, LuaFunctionJob(onStderr), userargs}
	} else {
		proc.Stderr = &outbuf
	}

	go func() {
		proc.Run()
		jobFunc := JobFunction{LuaFunctionJob(onExit), string(outbuf.Bytes()), userargs}
		jobs <- jobFunc
	}()

	return proc
}

func JobStop(cmd *exec.Cmd) {
	cmd.Process.Kill()
}

func JobSend(cmd *exec.Cmd, data string) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	stdin.Write([]byte(data))
}
