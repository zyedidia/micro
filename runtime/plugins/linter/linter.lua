function linter_lint(linter, cmd, errorformat)
    view:ClearGutterMessages(linter)

    local handle = io.popen(cmd)
    local lines = linter_split(handle:read("*a"), "\n")
    handle:close()

    messenger:Message(view.Buf.Path)
    local regex = errorformat:gsub("%%f", "(.+)"):gsub("%%l", "(%d+)"):gsub("%%m", "(.+)")
    for _,line in ipairs(lines) do
        if string.find(line, regex) then
            local file, line, msg = string.match(line, regex)
            if linter_basename(view.Buf.Path) == linter_basename(file) then
                view:GutterMessage(linter, tonumber(line), msg, 2)
            end
        end
    end
end

function linter_split(str, sep)
    local result = {}
    local regex = ("([^%s]+)"):format(sep)
    for each in str:gmatch(regex) do
        table.insert(result, each)
    end
    return result
end

function linter_basename(file)
    local name = string.gsub(file, "(.*/)(.*)", "%2")
    return name
end
