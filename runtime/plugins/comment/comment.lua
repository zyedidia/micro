VERSION = "1.0.0",

local util = import("micro/util")
local config = import("micro/config")
local buffer = import("micro/buffer")
local micro = import("micro")

local ft = {
    apacheconf = "# %s",
    batch = ":: %s",
    c = "// %s",
    c++ = "// %s",
    cmake = "# %s",
    conf = "# %s",
    crystal = "# %s",
    css = "/* %s */",
    d = "// %s",
    dart = "// %s",
    dockerfile = "# %s",
    elm = "-- %s",
    fish = "# %s",
    gdscript = "# %s",
    glsl = "// %s",
    go = "// %s",
    haskell = "-- %s",
    html = "<!-- %s -->",
    ini = "; %s",
    java = "// %s",
    javascript = "// %s",
    jinja2 = "{# %s #}",
    json = "// %s",
    julia = "# %s",
    kotlin = "// %s",
    lua = "-- %s",
    markdown = "<!-- %s -->",
    nginx = "# %s",
    nim = "# %s",
    objc = "// %s",
    ocaml = "(* %s *)",
    pascal = "{ %s }",
    perl = "# %s",
    php = "// %s",
    pony = "// %s",
    powershell = "# %s",
    proto = "// %s",
    python2 = "# %s",
    python3 = "# %s",
    python = "# %s",
    renpy = "# %s",
    ruby = "# %s",
    rust = "// %s",
    scala = "// %s",
    shell = "# %s",
    sql = "-- %s",
    swift = "// %s",
    tex = "% %s",
    toml = "# %s",
    twig = "{# %s #}",
    typescript = "// %s",
    v = "// %s",
    xml = "<!-- %s -->",
    yaml = "# %s",
    zig = "// %s",
    zscript = "// %s",
    zsh = "# %s"
}

function updateCommentType(buf)
    -- Using DoSetOptionNative to avoid `LocalSettings[option] = true`
    -- so that "comment.type" can be reset by a "filetype" change to default.
    if (buf.Settings["comment.type"] == "") then
        -- This won't get triggered if a filetype is change via `setlocal filetype`
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

function commentLine(bp, lineN, indentLen)
    updateCommentType(bp.Buf)

    local line = bp.Buf:Line(lineN)
    local commentType = bp.Buf.Settings["comment.type"]
    local sel = -bp.Cursor.CurSelection
    local curpos = -bp.Cursor.Loc
    local index = string.find(commentType, "%%s") - 1
    local indent = string.sub(line, 1, indentLen)
    local trimmedLine = string.sub(line, indentLen + 1)
    trimmedLine = trimmedLine:gsub("%%", "%%%%")
    local commentedLine = commentType:gsub("%%s", trimmedLine)
    bp.Buf:Replace(buffer.Loc(0, lineN), buffer.Loc(#line, lineN), indent .. commentedLine)
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
    bp.Cursor:StoreVisualX()
end

function uncommentLine(bp, lineN, commentRegex)
    updateCommentType(bp.Buf)

    local line = bp.Buf:Line(lineN)
    local commentType = bp.Buf.Settings["comment.type"]
    local sel = -bp.Cursor.CurSelection
    local curpos = -bp.Cursor.Loc
    local index = string.find(commentType, "%%s") - 1
    if not string.match(line, commentRegex) then
        commentRegex = commentRegex:gsub("%s+", "%s*")
    end
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
    bp.Cursor:StoreVisualX()
end

function toggleCommentLine(bp, lineN, commentRegex)
    if isCommented(bp, lineN, commentRegex) then
        uncommentLine(bp, lineN, commentRegex)
    else
        commentLine(bp, lineN, #util.GetLeadingWhitespace(bp.Buf:Line(lineN)))
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

    -- We assume that the indentation is either tabs only or spaces only
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
            uncommentLine(bp, line, commentRegex)
        else
            commentLine(bp, line, indentMin)
        end
    end
end

function comment(bp, args)
    updateCommentType(bp.Buf)

    local commentType = bp.Buf.Settings["comment.type"]
    local commentRegex = "^%s*" .. commentType:gsub("%%","%%%%"):gsub("%$","%$"):gsub("%)","%)"):gsub("%(","%("):gsub("%?","%?"):gsub("%*", "%*"):gsub("%-", "%-"):gsub("%.", "%."):gsub("%+", "%+"):gsub("%]", "%]"):gsub("%[", "%["):gsub("%%%%s", "(.*)")

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
