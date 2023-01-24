VERSION = "1.0.0"

local config = import("micro/config")

function removeline(bp, args)
	bp.Cursor.SelectLine(bp.Cursor)
	bp.Cursor.DeleteSelection(bp.Cursor)
end

function init()
    config.MakeCommand("removeline", removeline, config.NoComplete)
    config.TryBindKey("Ctrl-Delete", "lua:removeline.removeline", false)
    config.AddRuntimeFile("removeline", config.RTHelp, "help/removeline.md")
end

return true
