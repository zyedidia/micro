VERSION = "1.0.1"

local util = import("micro/util")
local config = import("micro/config")
local buffer = import("micro/buffer")
local micro = import("micro")

local ft = {}

ft["apacheconf"] = "# %s"
ft["batch"] = ":: %s"
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
ft["json"] = "// %s"
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
ft["typescript"] = "// %s"
ft["v"] = "// %s"
ft["xml"] = "<!-- %s -->"
ft["yaml"] = "# %s"
ft["zig"] = "// %s"
ft["zscript"] = "// %s"
ft["zsh"] = "# %s"

function updateCommentType(buf)
    -- NOTE: Using DoSetOptionNative to avoid LocalSettings[option] = true
    -- so that "comment.type" can be reset by a "filetype" change to default.
    if (buf.Settings["comment.type"] == "") then
        -- NOTE: This won't get triggered if a filetype is change via `setlocal filetype`
        -- since it is not registered with `RegisterGlobalOption()``
        if buf.Settings["commenttype"] ~= nil then
            buf:DoSetOptionNative("comment.type", buf.Settings["commenttype"])
        else
            if (ft[buf.Settings["filetype"]] ~= nil) then
                buf:DoSetOptionNative("comment.type", ft[buf.Settings["filetype"]])
            else
                buf:DoSetOptionNative("comment.type", "# %s")
            end
        end
    end
end

function isCommented(bp, lineN, commentRegex)
    local line = bp.Buf:Line(lineN)
    local regex = commentRegex:gsub("%s+", "%s*")
    if string.match(line, regex) then
        return true
    end
    return false
end

function commentLine(bp, cursor, lineN, indentLen)
    updateCommentType(bp.Buf)

    local line = bp.Buf:Line(lineN)
    local commentType = bp.Buf.Settings["comment.type"]
    local sel = -cursor.CurSelection
    local curpos = -cursor.Loc
    local index = string.find(commentType, "%%s") - 1
    local indent = string.sub(line, 1, indentLen)
    local trimmedLine = string.sub(line, indentLen + 1)
    trimmedLine = trimmedLine:gsub("%%", "%%%%")
    local commentedLine = commentType:gsub("%%s", trimmedLine)
    bp.Buf:Replace(buffer.Loc(0, lineN), buffer.Loc(#line, lineN), indent .. commentedLine)
    if cursor:HasSelection() then
        cursor.CurSelection[1].Y = sel[1].Y
        cursor.CurSelection[2].Y = sel[2].Y
        cursor.CurSelection[1].X = sel[1].X
        cursor.CurSelection[2].X = sel[2].X
    else
        cursor.X = curpos.X + index
        cursor.Y = curpos.Y
    end
    cursor:Relocate()
    cursor:StoreVisualX()
end

function uncommentLine(bp, cursor, lineN, commentRegex)
    updateCommentType(bp.Buf)

    local line = bp.Buf:Line(lineN)
    local commentType = bp.Buf.Settings["comment.type"]
    local sel = -cursor.CurSelection
    local curpos = -cursor.Loc
    local index = string.find(commentType, "%%s") - 1
    if not string.match(line, commentRegex) then
        commentRegex = commentRegex:gsub("%s+", "%s*")
    end
    if string.match(line, commentRegex) then
        uncommentedLine = string.match(line, commentRegex)
        bp.Buf:Replace(buffer.Loc(0, lineN), buffer.Loc(#line, lineN), util.GetLeadingWhitespace(line) .. uncommentedLine)
        if cursor:HasSelection() then
            cursor.CurSelection[1].Y = sel[1].Y
            cursor.CurSelection[2].Y = sel[2].Y
            cursor.CurSelection[1].X = sel[1].X
            cursor.CurSelection[2].X = sel[2].X
        else
            cursor.X = curpos.X - index
            cursor.Y = curpos.Y
        end
    end
    cursor:Relocate()
    cursor:StoreVisualX()
end

function toggleCommentLine(bp, cursor, lineN, commentRegex)
    if isCommented(bp, lineN, commentRegex) then
        uncommentLine(bp, cursor, lineN, commentRegex)
    else
        commentLine(bp, cursor, lineN, #util.GetLeadingWhitespace(bp.Buf:Line(lineN)))
    end
end

function toggleCommentSelection(bp, cursor, startLine, endLine, commentRegex)
    local allComments = true
    for line = startLine, endLine do
        if not isCommented(bp, line, commentRegex) then
            allComments = false
            break
        end
    end

    -- NOTE: we assume that the indentation is either tabs only or spaces only
    local indentMin = -1
    if not allComments then
        for line = startLine, endLine do
            local indentLen = #util.GetLeadingWhitespace(bp.Buf:Line(line))
            if indentMin == -1 or indentLen < indentMin then
                indentMin = indentLen
            end
        end
    end

    for line = startLine, endLine do
        if allComments then
            uncommentLine(bp, cursor, line, commentRegex)
        else
            commentLine(bp, cursor, line, indentMin)
        end
    end
end

function comment(bp, args)
    updateCommentType(bp.Buf)

	local cursors = bp.Buf:GetCursors()

    local commentType = bp.Buf.Settings["comment.type"]
    local commentRegex = "^%s*" .. commentType:gsub("%%","%%%%"):gsub("%$","%$"):gsub("%)","%)"):gsub("%(","%("):gsub("%?","%?"):gsub("%*", "%*"):gsub("%-", "%-"):gsub("%.", "%."):gsub("%+", "%+"):gsub("%]", "%]"):gsub("%[", "%["):gsub("%%%%s", "(.*)")

	for i = 1, #cursors do
		local cursor = cursors[i]
	    if cursor:HasSelection() then
	        if cursor.CurSelection[1]:GreaterThan(-cursor.CurSelection[2]) then
	            local endLine = cursor.CurSelection[1].Y
	            if cursor.CurSelection[1].X == 0 then
	                endLine = endLine - 1
	            end
	            toggleCommentSelection(bp, cursor, cursor.CurSelection[2].Y, endLine, commentRegex)
	        else
	            local endLine = cursor.CurSelection[2].Y
	            if cursor.CurSelection[2].X == 0 then
	                endLine = endLine - 1
	            end
	            toggleCommentSelection(bp, cursor, cursor.CurSelection[1].Y, endLine, commentRegex)
	        end
	    else
	        toggleCommentLine(bp, cursor, cursor.Y, commentRegex)
	    end
    end
end

function string.starts(String,Start)
    return string.sub(String,1,string.len(Start))==Start
end

function preinit()
    config.RegisterCommonOption("comment", "type", "")
end

function init()
    config.MakeCommand("comment", comment, config.NoComplete)
    config.TryBindKey("Alt-/", "lua:comment.comment", false)
    config.TryBindKey("CtrlUnderscore", "lua:comment.comment", false)
    config.AddRuntimeFile("comment", config.RTHelp, "help/comment.md")
end
