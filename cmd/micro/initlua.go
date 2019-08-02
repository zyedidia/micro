package main

import (
	"log"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/zyedidia/micro/internal/action"
	"github.com/zyedidia/micro/internal/display"
	ulua "github.com/zyedidia/micro/internal/lua"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/shell"
	"github.com/zyedidia/micro/internal/util"
)

func init() {
	ulua.L = lua.NewState()
	ulua.L.SetGlobal("import", luar.New(ulua.L, LuaImport))
}

func LuaImport(pkg string) *lua.LTable {
	switch pkg {
	case "micro":
		return luaImportMicro()
	case "micro/shell":
		return luaImportMicroShell()
	case "micro/util":
		return luaImportMicroUtil()
	default:
		return ulua.Import(pkg)
	}
}

func luaImportMicro() *lua.LTable {
	pkg := ulua.L.NewTable()

	ulua.L.SetField(pkg, "TermMessage", luar.New(ulua.L, screen.TermMessage))
	ulua.L.SetField(pkg, "TermError", luar.New(ulua.L, screen.TermError))
	ulua.L.SetField(pkg, "InfoBar", luar.New(ulua.L, action.GetInfoBar))
	ulua.L.SetField(pkg, "Log", luar.New(ulua.L, log.Println))
	ulua.L.SetField(pkg, "SetStatusInfoFn", luar.New(ulua.L, display.SetStatusInfoFnLua))
	// ulua.L.SetField(pkg, "TryBindKey", luar.New(ulua.L, action.TryBindKey))

	return pkg
}

func luaImportMicroShell() *lua.LTable {
	pkg := ulua.L.NewTable()

	ulua.L.SetField(pkg, "ExecCommand", luar.New(ulua.L, shell.ExecCommand))
	ulua.L.SetField(pkg, "RunCommand", luar.New(ulua.L, shell.RunCommand))
	ulua.L.SetField(pkg, "RunBackgroundShell", luar.New(ulua.L, shell.RunBackgroundShell))
	ulua.L.SetField(pkg, "RunInteractiveShell", luar.New(ulua.L, shell.RunInteractiveShell))

	return pkg
}

func luaImportMicroUtil() *lua.LTable {
	pkg := ulua.L.NewTable()

	ulua.L.SetField(pkg, "RuneAt", luar.New(ulua.L, util.LuaRuneAt))
	ulua.L.SetField(pkg, "GetLeadingWhitespace", luar.New(ulua.L, util.LuaGetLeadingWhitespace))
	ulua.L.SetField(pkg, "IsWordChar", luar.New(ulua.L, util.LuaIsWordChar))

	return pkg
}
