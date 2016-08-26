package main

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/layeh/gopher-luar"
	"github.com/yuin/gopher-lua"
)

var loadedPlugins []string

var preInstalledPlugins = []string{
	"go",
	"linter",
	"autoclose",
}

// Call calls the lua function 'function'
// If it does not exist nothing happens, if there is an error,
// the error is returned
func Call(function string, args ...interface{}) (lua.LValue, error) {
	var luaFunc lua.LValue
	if strings.Contains(function, ".") {
		plugin := L.GetGlobal(strings.Split(function, ".")[0])
		if plugin.String() == "nil" {
			return nil, errors.New("function does not exist: " + function)
		}
		luaFunc = L.GetField(plugin, strings.Split(function, ".")[1])
	} else {
		luaFunc = L.GetGlobal(function)
	}

	if luaFunc.String() == "nil" {
		return nil, errors.New("function does not exist: " + function)
	}
	var luaArgs []lua.LValue
	for _, v := range args {
		luaArgs = append(luaArgs, luar.New(L, v))
	}
	err := L.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    1,
		Protect: true,
	}, luaArgs...)
	ret := L.Get(-1) // returned value
	if ret.String() != "nil" {
		L.Pop(1) // remove received value
	}
	return ret, err
}

// LuaFunctionBinding is a function generator which takes the name of a lua function
// and creates a function that will call that lua function
// Specifically it creates a function that can be called as a binding because this is used
// to bind keys to lua functions
func LuaFunctionBinding(function string) func(*View, bool) bool {
	return func(v *View, _ bool) bool {
		_, err := Call(function, nil)
		if err != nil {
			TermMessage(err)
		}
		return false
	}
}

func unpack(old []string) []interface{} {
	new := make([]interface{}, len(old))
	for i, v := range old {
		new[i] = v
	}
	return new
}

// LuaFunctionCommand is the same as LuaFunctionBinding except it returns a normal function
// so that a command can be bound to a lua function
func LuaFunctionCommand(function string) func([]string) {
	return func(args []string) {
		_, err := Call(function, unpack(args)...)
		if err != nil {
			TermMessage(err)
		}
	}
}

func LuaFunctionJob(function string) func(string, ...string) {
	return func(output string, args ...string) {
		_, err := Call(function, unpack(append([]string{output}, args...))...)
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
		alreadyExists := false
		for _, pl := range loadedPlugins {
			if pl == pluginName {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
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

	if _, err := os.Stat(configDir + "/init.lua"); err == nil {
		pluginDef := "\nlocal P = {}\n" + "init" + " = P\nsetmetatable(" + "init" + ", {__index = _G})\nsetfenv(1, P)\n"
		data, _ := ioutil.ReadFile(configDir + "/init.lua")
		if err := L.DoString(pluginDef + string(data)); err != nil {
			TermMessage(err)
		}
		loadedPlugins = append(loadedPlugins, "init")
	}
}
