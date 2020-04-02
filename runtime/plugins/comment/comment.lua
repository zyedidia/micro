VERSION = "1.0.0"

local util = import("micro/util")
local config = import("micro/config")
local buffer = import("micro/buffer")

local ft = {}

ft["c"] = "// %s"
ft["c++"] = "// %s"
ft["go"] = "// %s"
ft["python"] = "# %s"
ft["python3"] = "# %s"
ft["html"] = "<!-- %s -->"
ft["java"] = "// %s"
ft["julia"] = "# %s"
ft["perl"] = "# %s"
ft["php"] = "// %s"
ft["rust"] = "// %s"
ft["shell"] = "# %s"
ft["lua"] = "-- %s"
ft["javascript"] = "// %s"
ft["ruby"] = "# %s"
ft["d"] = "// %s"
ft["swift"] = "// %s"
ft["elm"] = "-- %s"

function onBufferOpen(buf)
    if buf.Settings["commenttype"] == nil then
        if ft[buf.Settings["filetype"]] ~= nil then
            buf.Settings["commenttype"] = ft[buf.Settings["filetype"]]
        else
            buf.Settings["commenttype"] = "# %s"
        end
    end
end

function commentLine(bp, lineN)
    local line = bp.Buf:Line(lineN)
    local commentType = bp.Buf.Settings["commenttype"]
    local commentRegex = "^%s*" .. commentType:gsub("%*", "%*"):gsub("%-", "%-"):gsub("%.", "%."):gsub("%+", "%+"):gsub("%]", "%]"):gsub("%[", "%["):gsub("%%s", "(.*)")
    local sel = -bp.Cursor.CurSelection
    local curpos = -bp.Cursor.Loc
    local index = string.find(commentType, "%%s") - 1
    if string.match(line, commentRegex) then
        uncommentedLine = string.match(line, commentRegex)
        bp.Buf:Replace(buffer.Loc(0, lineN), buffer.Loc(#line, lineN), util.GetLeadingWhitespace(line) .. uncommentedLine)
        if bp.Cursor:HasSelection() then
            bp.Cursor.CurSelection[1].Y = sel[1].Y
            bp.Cursor.CurSelection[2].Y = sel[2].Y
            bp.Cursor.CurSelection[1].X = sel[1].X
            bp.Cursor.CurSelection[2].X = sel[2].X
        else
            bp.Cursor.X = curpos.X - index
            bp.Cursor.Y = curpos.Y
        end
    else
        local commentedLine = commentType:gsub("%%s", trim(line))
        bp.Buf:Replace(buffer.Loc(0, lineN), buffer.Loc(#line, lineN), util.GetLeadingWhitespace(line) .. commentedLine)
        if bp.Cursor:HasSelection() then
            bp.Cursor.CurSelection[1].Y = sel[1].Y
            bp.Cursor.CurSelection[2].Y = sel[2].Y
            bp.Cursor.CurSelection[1].X = sel[1].X
            bp.Cursor.CurSelection[2].X = sel[2].X
        else
            bp.Cursor.X = curpos.X + index
            bp.Cursor.Y = curpos.Y
        end
    end
    bp.Cursor:Relocate()
    bp.Cursor.LastVisualX = bp.Cursor:GetVisualX()
end

function commentSelection(bp, startLine, endLine)
    for line = startLine, endLine do
        commentLine(bp, line)
    end
end

function comment(bp, args)
    if bp.Cursor:HasSelection() then
        if bp.Cursor.CurSelection[1]:GreaterThan(-bp.Cursor.CurSelection[2]) then
            local endLine = bp.Cursor.CurSelection[1].Y
            if bp.Cursor.CurSelection[1].X == 0 then
                endLine = endLine - 1
            end
            commentSelection(bp, bp.Cursor.CurSelection[2].Y, endLine)
        else
            local endLine = bp.Cursor.CurSelection[2].Y
            if bp.Cursor.CurSelection[2].X == 0 then
                endLine = endLine - 1
            end
            commentSelection(bp, bp.Cursor.CurSelection[1].Y, endLine)
        end
    else
        commentLine(bp, bp.Cursor.Y)
    end
end

function trim(s)
    return (s:gsub("^%s*(.-)%s*$", "%1"))
end

function string.starts(String,Start)
    return string.sub(String,1,string.len(Start))==Start
end

function init()
    config.MakeCommand("comment", comment, config.NoComplete)
    config.TryBindKey("Alt-/", "lua:comment.comment", false)
    config.AddRuntimeFile("comment", config.RTHelp, "help/comment.md")
end
