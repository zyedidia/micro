package main

import (
	"log"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/zyedidia/micro/internal/action"
	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
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
	case "micro/config":
		return luaImportMicroConfig()
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

	return pkg
}

func luaImportMicroConfig() *lua.LTable {
	pkg := ulua.L.NewTable()

	ulua.L.SetField(pkg, "MakeCommand", luar.New(ulua.L, action.LuaMakeCommand))
	ulua.L.SetField(pkg, "FileComplete", luar.New(ulua.L, buffer.FileComplete))
	ulua.L.SetField(pkg, "HelpComplete", luar.New(ulua.L, action.HelpComplete))
	ulua.L.SetField(pkg, "OptionComplete", luar.New(ulua.L, action.OptionComplete))
	ulua.L.SetField(pkg, "OptionValueComplete", luar.New(ulua.L, action.OptionValueComplete))
	ulua.L.SetField(pkg, "NoComplete", luar.New(ulua.L, nil))
	ulua.L.SetField(pkg, "TryBindKey", luar.New(ulua.L, action.TryBindKey))
	ulua.L.SetField(pkg, "Reload", luar.New(ulua.L, action.ReloadConfig))
	ulua.L.SetField(pkg, "AddRuntimeFileFromMemory", luar.New(ulua.L, config.PluginAddRuntimeFileFromMemory))
	ulua.L.SetField(pkg, "AddRuntimeFilesFromDirectory", luar.New(ulua.L, config.PluginAddRuntimeFileFromMemory))
	ulua.L.SetField(pkg, "AddRuntimeFile", luar.New(ulua.L, config.PluginAddRuntimeFile))
	ulua.L.SetField(pkg, "ListRuntimeFiles", luar.New(ulua.L, config.PluginListRuntimeFiles))
	ulua.L.SetField(pkg, "ReadRuntimeFile", luar.New(ulua.L, config.PluginReadRuntimeFile))
	ulua.L.SetField(pkg, "RTColorscheme", luar.New(ulua.L, config.RTColorscheme))
	ulua.L.SetField(pkg, "RTSyntax", luar.New(ulua.L, config.RTSyntax))
	ulua.L.SetField(pkg, "RTHelp", luar.New(ulua.L, config.RTHelp))
	ulua.L.SetField(pkg, "RTPlugin", luar.New(ulua.L, config.RTPlugin))

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
	ulua.L.SetField(pkg, "Loc", luar.New(ulua.L, func(x, y int) buffer.Loc {
		return buffer.Loc{x, y}
	}))

	return pkg
}

func luaImportMicroUtil() *lua.LTable {
	pkg := ulua.L.NewTable()

	ulua.L.SetField(pkg, "RuneAt", luar.New(ulua.L, util.LuaRuneAt))
	ulua.L.SetField(pkg, "GetLeadingWhitespace", luar.New(ulua.L, util.LuaGetLeadingWhitespace))
	ulua.L.SetField(pkg, "IsWordChar", luar.New(ulua.L, util.LuaIsWordChar))

	return pkg
}
