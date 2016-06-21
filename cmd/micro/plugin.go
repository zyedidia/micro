package main

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/layeh/gopher-luar"
	"github.com/yuin/gopher-lua"
)

var loadedPlugins []string

var preInstalledPlugins = []string{
	"go",
	"linter",
}

// Call calls the lua function 'function'
// If it does not exist nothing happens, if there is an error,
// the error is returned
func Call(function string, args []string) error {
	var luaFunc lua.LValue
	if strings.Contains(function, ".") {
		plugin := L.GetGlobal(strings.Split(function, ".")[0])
		if plugin.String() == "nil" {
			return errors.New("function does not exist: " + function)
		}
		luaFunc = L.GetField(plugin, strings.Split(function, ".")[1])
	} else {
		luaFunc = L.GetGlobal(function)
	}

	if luaFunc.String() == "nil" {
		return errors.New("function does not exist: " + function)
	}
	var luaArgs []lua.LValue
	for _, v := range args {
		luaArgs = append(luaArgs, luar.New(L, v))
	}
	err := L.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    0,
		Protect: true,
	}, luaArgs...)
	return err
}

// LuaFunctionBinding is a function generator which takes the name of a lua function
// and creates a function that will call that lua function
// Specifically it creates a function that can be called as a binding because this is used
// to bind keys to lua functions
func LuaFunctionBinding(function string) func(*View) bool {
	return func(v *View) bool {
		err := Call(function, nil)
		if err != nil {
			TermMessage(err)
		}
		return false
	}
}

// LuaFunctionCommand is the same as LuaFunctionBinding except it returns a normal function
// so that a command can be bound to a lua function
func LuaFunctionCommand(function string) func([]string) {
	return func(args []string) {
		err := Call(function, args)
		if err != nil {
			TermMessage(err)
		}
	}
}

func LuaFunctionJob(function string) func(string, ...string) {
	return func(output string, args ...string) {
		err := Call(function, append([]string{output}, args...))
		if err != nil {
			TermMessage(err)
		}
	}
}

// LoadPlugins loads the pre-installed plugins and the plugins located in ~/.config/micro/plugins
func LoadPlugins() {
	files, _ := ioutil.ReadDir(configDir + "/plugins")
	for _, plugin := range files {
		if plugin.IsDir() {
			pluginName := plugin.Name()
			files, _ := ioutil.ReadDir(configDir + "/plugins/" + pluginName)
			for _, f := range files {
				if f.Name() == pluginName+".lua" {
					data, _ := ioutil.ReadFile(configDir + "/plugins/" + pluginName + "/" + f.Name())
					pluginDef := "\nlocal P = {}\n" + pluginName + " = P\nsetmetatable(" + pluginName + ", {__index = _G})\nsetfenv(1, P)\n"

					if err := L.DoString(pluginDef + string(data)); err != nil {
						TermMessage(err)
						continue
					}
					loadedPlugins = append(loadedPlugins, pluginName)
				}
			}
		}
	}

	for _, pluginName := range preInstalledPlugins {
		plugin := "runtime/plugins/" + pluginName + "/" + pluginName + ".lua"
		data, err := Asset(plugin)
		if err != nil {
			TermMessage("Error loading pre-installed plugin: " + pluginName)
			continue
		}
		pluginDef := "\nlocal P = {}\n" + pluginName + " = P\nsetmetatable(" + pluginName + ", {__index = _G})\nsetfenv(1, P)\n"
		if err := L.DoString(pluginDef + string(data)); err != nil {
			TermMessage(err)
			continue
		}
		loadedPlugins = append(loadedPlugins, pluginName)
	}
}
