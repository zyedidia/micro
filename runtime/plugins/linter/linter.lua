if GetOption("linter") == nil then
    AddOption("linter", true)
end

function linter_onSave()
    if GetOption("linter") then
        local ft = CurView().Buf.FileType
        local file = CurView().Buf.Path
        local devnull = "/dev/null"
        if OS == "windows" then
            devnull = "NUL"
        end
        if ft == "Go" then
            linter_lint("gobuild", "go build -o " .. devnull, "%f:%l: %m")
            linter_lint("golint", "golint " .. CurView().Buf.Path, "%f:%l:%d+: %m")
        elseif ft == "Lua" then
            linter_lint("luacheck", "luacheck --no-color " .. file, "%f:%l:%d+: %m")
        elseif ft == "Python" then
            linter_lint("pyflakes", "pyflakes " .. file, "%f:%l: %m")
        elseif ft == "C" then
            linter_lint("gcc", "gcc -fsyntax-only -Wall -Wextra " .. file, "%f:%l:%d+:.+: %m")
        elseif ft == "D" then
            linter_lint("dmd", "dmd -color=off -o- -w -wi -c " .. file, "%f%(%l%):.+: %m")
        elseif ft == "Java" then
            linter_lint("javac", "javac " .. file, "%f:%l: error: %m")
        elseif ft == "JavaScript" then
            linter_lint("jshint", "jshint " .. file, "%f: line %l,.+, %m")
        end
    else
        CurView():ClearAllGutterMessages()
    end
end

function linter_lint(linter, cmd, errorformat)
    CurView():ClearGutterMessages(linter)

    local handle = io.popen("(" .. cmd .. ")" .. " 2>&1")
    local lines = linter_split(handle:read("*a"), "\n")
    handle:close()

    local regex = errorformat:gsub("%%f", "(.+)"):gsub("%%l", "(%d+)"):gsub("%%m", "(.+)")
    for _,line in ipairs(lines) do
        -- Trim whitespace
        line = line:match("^%s*(.+)%s*$")
        if string.find(line, regex) then
            local file, line, msg = string.match(line, regex)
            if linter_basename(CurView().Buf.Path) == linter_basename(file) then
                CurView():GutterMessage(linter, tonumber(line), msg, 2)
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
