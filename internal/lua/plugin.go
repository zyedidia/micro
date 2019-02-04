package lua

import (
	"errors"
	"io/ioutil"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

var ErrNoSuchFunction = errors.New("No such function exists")

type Plugin struct {
	name  string
	files []string
}

func NewPluginFromDir(name string, dir string) (*Plugin, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	p := new(Plugin)
	p.name = name

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".lua") {
			p.files = append(p.files, dir+f.Name())
		}
	}

	return p, nil
}

func (p *Plugin) Load() error {
	for _, f := range p.files {
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		err = LoadFile(p.name, f, dat)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugin) Call(fn string, args ...lua.LValue) (lua.LValue, error) {
	plug := L.GetGlobal(p.name)
	luafn := L.GetField(plug, fn)
	if luafn == lua.LNil {
		return nil, ErrNoSuchFunction
	}
	err := L.CallByParam(lua.P{
		Fn:      luafn,
		NRet:    1,
		Protect: true,
	}, args...)
	if err != nil {
		return nil, err
	}
	ret := L.Get(-1)
	L.Pop(1)
	return ret, nil
}
