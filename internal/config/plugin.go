package config

import (
	"errors"

	lua "github.com/yuin/gopher-lua"
	ulua "github.com/zyedidia/micro/internal/lua"
)

var ErrNoSuchFunction = errors.New("No such function exists")

// LoadAllPlugins loads all detected plugins (in runtime/plugins and ConfigDir/plugins)
func LoadAllPlugins() {
	for _, p := range Plugins {
		p.Load()
	}
}

// RunPluginFn runs a given function in all plugins
// returns an error if any of the plugins had an error
func RunPluginFn(fn string, args ...lua.LValue) error {
	var reterr error
	for _, p := range Plugins {
		if !p.IsEnabled() {
			continue
		}
		_, err := p.Call(fn, args...)
		if err != nil && err != ErrNoSuchFunction {
			reterr = errors.New("Plugin " + p.Name + ": " + err.Error())
		}
	}
	return reterr
}

// RunPluginFnBool runs a function in all plugins and returns
// false if any one of them returned false
// also returns an error if any of the plugins had an error
func RunPluginFnBool(fn string, args ...lua.LValue) (bool, error) {
	var reterr error
	retbool := true
	for _, p := range Plugins {
		if !p.IsEnabled() {
			continue
		}
		val, err := p.Call(fn, args...)
		if err == ErrNoSuchFunction {
			continue
		}
		if err != nil {
			reterr = errors.New("Plugin " + p.Name + ": " + err.Error())
			continue
		}
		if v, ok := val.(lua.LBool); !ok {
			reterr = errors.New(p.Name + "." + fn + " should return a boolean")
		} else {
			retbool = retbool && bool(v)
		}
	}
	return retbool, reterr
}

type Plugin struct {
	Name   string        // name of plugin
	Info   RuntimeFile   // json file containing info
	Srcs   []RuntimeFile // lua files
	Loaded bool
}

func (p *Plugin) IsEnabled() bool {
	if v, ok := GlobalSettings[p.Name]; ok {
		return v.(bool)
	}
	return true
}

var Plugins []*Plugin

func (p *Plugin) Load() error {
	for _, f := range p.Srcs {
		if !p.IsEnabled() {
			return nil
		}
		dat, err := f.Data()
		if err != nil {
			return err
		}
		err = ulua.LoadFile(p.Name, f.Name(), dat)
		if err != nil {
			return err
		}
		p.Loaded = true
		if _, ok := GlobalSettings[p.Name]; !ok {
			AddOption(p.Name, true)
		}
	}
	return nil
}

func (p *Plugin) Call(fn string, args ...lua.LValue) (lua.LValue, error) {
	plug := ulua.L.GetGlobal(p.Name)
	luafn := ulua.L.GetField(plug, fn)
	if luafn == lua.LNil {
		return nil, ErrNoSuchFunction
	}
	err := ulua.L.CallByParam(lua.P{
		Fn:      luafn,
		NRet:    1,
		Protect: true,
	}, args...)
	if err != nil {
		return nil, err
	}
	ret := ulua.L.Get(-1)
	ulua.L.Pop(1)
	return ret, nil
}
