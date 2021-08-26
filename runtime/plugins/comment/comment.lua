VERSION = "1.0.0"

local util = import("micro/util")
local config = import("micro/config")
local buffer = import("micro/buffer")

local ft = {}

ft["apacheconf"] = "# %s"
ft["bat"] = ":: %s"
ft["c"] = "// %s"
ft["c++"] = "// %s"
ft["cmake"] = "# %s"
ft["conf"] = "# %s"
ft["crystal"] = "# %s"
ft["css"] = "/* %s */"
ft["d"] = "// %s"
ft["dart"] = "// %s"
ft["dockerfile"] = "# %s"
ft["elm"] = "-- %s"
ft["fish"] = "# %s"
ft["gdscript"] = "# %s"
ft["glsl"] = "// %s"
ft["go"] = "// %s"
ft["haskell"] = "-- %s"
ft["html"] = "<!-- %s -->"
ft["ini"] = "; %s"
ft["java"] = "// %s"
ft["javascript"] = "// %s"
ft["jinja2"] = "{# %s #}"
ft["julia"] = "# %s"
ft["kotlin"] = "// %s"
ft["lua"] = "-- %s"
ft["markdown"] = "<!-- %s -->"
ft["nginx"] = "# %s"
ft["nim"] = "# %s"
ft["objc"] = "// %s"
ft["ocaml"] = "(* %s *)"
ft["pascal"] = "{ %s }"
ft["perl"] = "# %s"
ft["php"] = "// %s"
ft["pony"] = "// %s"
ft["powershell"] = "# %s"
ft["proto"] = "// %s"
ft["python"] = "# %s"
ft["python3"] = "# %s"
ft["ruby"] = "# %s"
ft["rust"] = "// %s"
ft["scala"] = "// %s"
ft["shell"] = "# %s"
ft["sql"] = "-- %s"
ft["swift"] = "// %s"
ft["tex"] = "% %s"
ft["toml"] = "# %s"
ft["twig"] = "{# %s #}"
ft["v"] = "// %s"
ft["xml"] = "<!-- %s -->"
ft["yaml"] = "# %s"
ft["zig"] = "// %s"
ft["zscript"] = "// %s"
ft["zsh"] = "# %s"

function updateCommentType(buf)
    if ft[buf.Settings["filetype"]] ~= nil and ft[buf.Settings["filetype"]] ~= nil then
        buf.Settings["commenttype"] = ft[buf.Settings["filetype"]]
    elseif buf.Settings["commenttype"] == nil then
        buf.Settings["commenttype"] = "# %s"
    end
end

function isCommented(bp, lineN, commentRegex)
    local line = bp.Buf:Line(lineN)
    if string.match(line, commentRegex) then
        return true
    end
    return false
end

function commentLine(bp, lineN)
    updateCommentType(bp.Buf)

    local line = bp.Buf:Line(lineN)
    local commentType = bp.Buf.Settings["commenttype"]
    local sel = -bp.Cursor.CurSelection
    local curpos = -bp.Cursor.Loc
    local index = string.find(commentType, "%%s") - 1
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
    bp.Cursor:Relocate()
    bp.Cursor.LastVisualX = bp.Cursor:GetVisualX()
end

function uncommentLine(bp, lineN, commentRegex)
    updateCommentType(bp.Buf)

    local line = bp.Buf:Line(lineN)
    local commentType = bp.Buf.Settings["commenttype"]
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
    end
    bp.Cursor:Relocate()
    bp.Cursor.LastVisualX = bp.Cursor:GetVisualX()
end

function toggleCommentLine(bp, lineN, commentRegex)
    if isCommented(bp, lineN, commentRegex) then
        uncommentLine(bp, lineN, commentRegex)
    else
        commentLine(bp, lineN)
    end
end

function toggleCommentSelection(bp, startLine, endLine, commentRegex)
    local allComments = true
    for line = startLine, endLine do
        if not isCommented(bp, line, commentRegex) then
            allComments = false
            break
        end
    end

    for line = startLine, endLine do
        if allComments then
            uncommentLine(bp, line, commentRegex)
        else
            commentLine(bp, line)
        end
    end
end

function comment(bp, args)
    updateCommentType(bp.Buf)

    local commentType = bp.Buf.Settings["commenttype"]
    local commentRegex = "^%s*" .. commentType:gsub("%%","%%%%"):gsub("%$","%$"):gsub("%)","%)"):gsub("%(","%("):gsub("%?","%?"):gsub("%*", "%*"):gsub("%-", "%-"):gsub("%.", "%."):gsub("%+", "%+"):gsub("%]", "%]"):gsub("%[", "%["):gsub("%%%%s", "(.*)"):gsub("%s+", "%s*")

    if bp.Cursor:HasSelection() then
        if bp.Cursor.CurSelection[1]:GreaterThan(-bp.Cursor.CurSelection[2]) then
            local endLine = bp.Cursor.CurSelection[1].Y
            if bp.Cursor.CurSelection[1].X == 0 then
                endLine = endLine - 1
            end
            toggleCommentSelection(bp, bp.Cursor.CurSelection[2].Y, endLine, commentRegex)
        else
            local endLine = bp.Cursor.CurSelection[2].Y
            if bp.Cursor.CurSelection[2].X == 0 then
                endLine = endLine - 1
            end
            toggleCommentSelection(bp, bp.Cursor.CurSelection[1].Y, endLine, commentRegex)
        end
    else
        toggleCommentLine(bp, bp.Cursor.Y, commentRegex)
    end
end

function trim(s)
    local trimmed = s:gsub("^%s*(.-)%s*$", "%1"):gsub("%%","%%%%")
    return trimmed
end

function string.starts(String,Start)
    return string.sub(String,1,string.len(Start))==Start
end

function init()
    config.MakeCommand("comment", comment, config.NoComplete)
    config.TryBindKey("Alt-/", "lua:comment.comment", false)
    config.TryBindKey("CtrlUnderscore", "lua:comment.comment", false)
    config.AddRuntimeFile("comment", config.RTHelp, "help/comment.md")
end
