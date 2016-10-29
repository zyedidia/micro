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

// LuaAction runs a Lua function as an action
func LuaAction(function string) bool {
	_, err := Call(function, nil)
	if err != nil {
		TermMessage(err)
	}
	return err == nil
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

// LuaFunctionComplete returns a function which can be used for autocomplete in plugins
func LuaFunctionComplete(function string) func(string) []string {
	return func(input string) (result []string) {

		res, err := Call(function, input)
		if err != nil {
			TermMessage(err)
		}
		if tbl, ok := res.(*lua.LTable); !ok {
			TermMessage(function, "should return a table of strings")
		} else {
			for i := 1; i <= tbl.Len(); i++ {
				val := tbl.RawGetInt(i)
				if v, ok := val.(lua.LString); !ok {
					TermMessage(function, "should return a table of strings")
				} else {
					result = append(result, string(v))
				}
			}
		}
		return result
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
	for _, plugin := range ListRuntimeFiles(RTPlugin) {
		alreadyExists := false
		pluginName := plugin.Name()
		for _, pl := range loadedPlugins {
			if pl == pluginName {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			data, err := plugin.Data()
			if err != nil {
				TermMessage("Error loading plugin: " + pluginName)
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
