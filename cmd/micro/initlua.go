package main

import (
	"log"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/zyedidia/micro/internal/action"
	"github.com/zyedidia/micro/internal/buffer"
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
	case "micro/buffer":
		return luaImportMicroBuffer()
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
	ulua.L.SetField(pkg, "JobStart", luar.New(ulua.L, shell.JobStart))
	ulua.L.SetField(pkg, "JobSpawn", luar.New(ulua.L, shell.JobSpawn))
	ulua.L.SetField(pkg, "JobStop", luar.New(ulua.L, shell.JobStop))
	ulua.L.SetField(pkg, "JobSend", luar.New(ulua.L, shell.JobSend))

	return pkg
}

func luaImportMicroBuffer() *lua.LTable {
	pkg := ulua.L.NewTable()

	ulua.L.SetField(pkg, "NewMessage", luar.New(ulua.L, buffer.NewMessage))
	ulua.L.SetField(pkg, "NewMessageAtLine", luar.New(ulua.L, buffer.NewMessageAtLine))
	ulua.L.SetField(pkg, "MTInfo", luar.New(ulua.L, buffer.MTInfo))
	ulua.L.SetField(pkg, "MTWarning", luar.New(ulua.L, buffer.MTWarning))
	ulua.L.SetField(pkg, "MTError", luar.New(ulua.L, buffer.MTError))

	return pkg
}

func luaImportMicroUtil() *lua.LTable {
	pkg := ulua.L.NewTable()

	ulua.L.SetField(pkg, "RuneAt", luar.New(ulua.L, util.LuaRuneAt))
	ulua.L.SetField(pkg, "GetLeadingWhitespace", luar.New(ulua.L, util.LuaGetLeadingWhitespace))
	ulua.L.SetField(pkg, "IsWordChar", luar.New(ulua.L, util.LuaIsWordChar))

	return pkg
}
