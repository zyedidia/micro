package shell

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	luar "layeh.com/gopher-luar"

	lua "github.com/yuin/gopher-lua"
	"github.com/zyedidia/micro/internal/config"
	ulua "github.com/zyedidia/micro/internal/lua"
	"github.com/zyedidia/micro/internal/screen"
)

var Jobs chan JobFunction

func init() {
	Jobs = make(chan JobFunction, 100)
}

// Jobs are the way plugins can run processes in the background
// A job is simply a process that gets executed asynchronously
// There are callbacks for when the job exits, when the job creates stdout
// and when the job creates stderr

// These jobs run in a separate goroutine but the lua callbacks need to be
// executed in the main thread (where the Lua VM is running) so they are
// put into the jobs channel which gets read by the main loop

// JobFunction is a representation of a job (this data structure is what is loaded
// into the jobs channel)
type JobFunction struct {
	Function func(string, ...interface{})
	Output   string
	Args     []interface{}
}

// A CallbackFile is the data structure that makes it possible to catch stderr and stdout write events
type CallbackFile struct {
	io.Writer

	callback func(string, ...interface{})
	args     []interface{}
}

func (f *CallbackFile) Write(data []byte) (int, error) {
	// This is either stderr or stdout
	// In either case we create a new job function callback and put it in the jobs channel
	jobFunc := JobFunction{f.callback, string(data), f.args}
	Jobs <- jobFunc
	return f.Writer.Write(data)
}

// JobStart starts a shell command in the background with the given callbacks
// It returns an *exec.Cmd as the job id
func JobStart(cmd string, onStdout, onStderr, onExit string, userargs ...interface{}) *exec.Cmd {
	return JobSpawn("sh", []string{"-c", cmd}, onStdout, onStderr, onExit, userargs...)
}

// JobSpawn starts a process with args in the background with the given callbacks
// It returns an *exec.Cmd as the job id
func JobSpawn(cmdName string, cmdArgs []string, onStdout, onStderr, onExit string, userargs ...interface{}) *exec.Cmd {
	// Set up everything correctly if the functions have been provided
	proc := exec.Command(cmdName, cmdArgs...)
	var outbuf bytes.Buffer
	if onStdout != "" {
		proc.Stdout = &CallbackFile{&outbuf, luaFunctionJob(onStdout), userargs}
	} else {
		proc.Stdout = &outbuf
	}
	if onStderr != "" {
		proc.Stderr = &CallbackFile{&outbuf, luaFunctionJob(onStderr), userargs}
	} else {
		proc.Stderr = &outbuf
	}

	go func() {
		// Run the process in the background and create the onExit callback
		proc.Run()
		jobFunc := JobFunction{luaFunctionJob(onExit), string(outbuf.Bytes()), userargs}
		Jobs <- jobFunc
	}()

	return proc
}

// JobStop kills a job
func JobStop(cmd *exec.Cmd) {
	cmd.Process.Kill()
}

// JobSend sends the given data into the job's stdin stream
func JobSend(cmd *exec.Cmd, data string) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	stdin.Write([]byte(data))
}

// luaFunctionJob returns a function that will call the given lua function
// structured as a job call i.e. the job output and arguments are provided
// to the lua function
func luaFunctionJob(fn string) func(string, ...interface{}) {
	luaFn := strings.Split(fn, ".")
	plName, plFn := luaFn[0], luaFn[1]
	pl := config.FindPlugin(plName)
	return func(output string, args ...interface{}) {
		var luaArgs []lua.LValue
		luaArgs = append(luaArgs, luar.New(ulua.L, output))
		for _, v := range args {
			luaArgs = append(luaArgs, luar.New(ulua.L, v))
		}
		_, err := pl.Call(plFn, luaArgs...)
		if err != nil && err != config.ErrNoSuchFunction {
			screen.TermMessage(err)
		}
	}
}
