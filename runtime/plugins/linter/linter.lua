local runtime = import("runtime")
local filepath = import("path/filepath")
local shell = import("micro/shell")
local buffer = import("micro/buffer")
local micro = import("micro")

local linters = {}

-- creates a linter entry, call from within an initialization function, not
-- directly at initial load time
--
-- name: name of the linter
-- filetype: filetype to check for to use linter
-- cmd: main linter process that is executed
-- args: arguments to pass to the linter process
--     use %f to refer to the current file name
--     use %d to refer to the current directory name
-- errorformat: how to parse the linter/compiler process output
--     %f: file, %l: line number, %m: error/warning message
-- os: list of OSs this linter is supported or unsupported on
--     optional param, default: []
-- whitelist: should the OS list be a blacklist (do not run the linter for these OSs)
-- or a whitelist (only run the linter for these OSs)
--     optional param, default: false (should blacklist)
-- domatch: should the filetype be interpreted as a lua pattern to match with
-- the actual filetype, or should the linter only activate on an exact match
--     optional param, default: false (require exact match)
function makeLinter(name, filetype, cmd, args, errorformat, os, whitelist, domatch)
    if linters[name] == nil then
        linters[name] = {}
        linters[name].filetype = filetype
        linters[name].cmd = cmd
        linters[name].args = args
        linters[name].errorformat = errorformat
        linters[name].os = os or {}
        linters[name].whitelist = whitelist or false
        linters[name].domatch = domatch or false
    end
end

function removeLinter(name)
    linters[name] = nil
end

function init()
    local devnull = "/dev/null"
    if runtime.GOOS == "windows" then
        devnull = "NUL"
    end

    makeLinter("gcc", "c", "gcc", {"-fsyntax-only", "-Wall", "-Wextra", "%f"}, "%f:%l:%d+:.+: %m")
    makeLinter("gcc", "c++", "gcc", {"-fsyntax-only","-std=c++14", "-Wall", "-Wextra", "%f"}, "%f:%l:%d+:.+: %m")
    makeLinter("dmd", "d", "dmd", {"-color=off", "-o-", "-w", "-wi", "-c", "%f"}, "%f%(%l%):.+: %m")
    makeLinter("gobuild", "go", "go", {"build", "-o", devnull}, "%f:%l: %m")
    makeLinter("golint", "go", "golint", {"%f"}, "%f:%l:%d+: %m")
    makeLinter("javac", "java", "javac", {"-d", "%d", "%f"}, "%f:%l: error: %m")
    makeLinter("jshint", "javascript", "jshint", {"%f"}, "%f: line %l,.+, %m")
    makeLinter("literate", "literate", "lit", {"-c", "%f"}, "%f:%l:%m", {}, false, true)
    makeLinter("luacheck", "lua", "luacheck", {"--no-color", "%f"}, "%f:%l:%d+: %m")
    makeLinter("nim", "nim", "nim", {"check", "--listFullPaths", "--stdout", "--hints:off", "%f"}, "%f.%l, %d+. %m")
    makeLinter("clang", "objective-c", "xcrun", {"clang", "-fsyntax-only", "-Wall", "-Wextra", "%f"}, "%f:%l:%d+:.+: %m")
    makeLinter("pyflakes", "python", "pyflakes", {"%f"}, "%f:%l:.-:? %m")
    makeLinter("mypy", "python", "mypy", {"%f"}, "%f:%l: %m")
    makeLinter("pylint", "python", "pylint", {"--output-format=parseable", "--reports=no", "%f"}, "%f:%l: %m")
    makeLinter("shfmt", "shell", "shfmt", {"%f"}, "%f:%l:%d+: %m")
    makeLinter("switfc", "swift", "xcrun", {"swiftc", "%f"}, "%f:%l:%d+:.+: %m", {"darwin"}, true)
    makeLinter("switfc", "swiftc", {"%f"}, "%f:%l:%d+:.+: %m", {"linux"}, true)
    makeLinter("yaml", "yaml", "yamllint", {"--format", "parsable", "%f"}, "%f:%l:%d+:.+ %m")
end

function contains(list, element)
    for k, v in pairs(list) do
        if v == element then
            return true
        end
    end
    return false
end

function runLinter(buf)
    local ft = buf:FileType()
    local file = buf.Path
    local dir = filepath.Dir(file)

    for k, v in pairs(linters) do
        local ftmatch = ft == v.filetype
        if v.domatch then
            ftmatch = string.match(ft, v.filetype)
        end

        local hasOS = contains(v.os, runtime.GOOS)
        if not hasOS and v.whitelist then
            ftmatch = false
        end
        if hasOS and not v.whitelist then
            ftmatch = false
        end

        for k, arg in pairs(v.args) do
            v.args[k] = arg:gsub("%%f", file):gsub("%%d", dir)
        end

        if ftmatch then
            lint(buf, k, v.cmd, v.args, v.errorformat)
        end
    end
end

function onSave(bp)
    micro.Log("SAVE")
    runLinter(bp.Buf)
    return false
end

function lint(buf, linter, cmd, args, errorformat)
    buf:ClearMessages("linter")

    shell.JobSpawn(cmd, args, "", "", "linter.onExit", buf, linter, errorformat)
end

function onExit(output, buf, linter, errorformat)
    micro.Log("ONEXIT")
    micro.Log(output)
    local lines = split(output, "\n")

    local regex = errorformat:gsub("%%f", "(..-)"):gsub("%%l", "(%d+)"):gsub("%%m", "(.+)")
    for _,line in ipairs(lines) do
        -- Trim whitespace
        line = line:match("^%s*(.+)%s*$")
        if string.find(line, regex) then
            local file, line, msg = string.match(line, regex)
            if basename(buf.Path) == basename(file) then
                local bmsg = buffer.NewMessageAtLine("linter", msg, tonumber(line), buffer.MTError)
                buf:AddMessage(bmsg)
            end
        end
    end
end

function split(str, sep)
    local result = {}
    local regex = ("([^%s]+)"):format(sep)
    for each in str:gmatch(regex) do
        table.insert(result, each)
    end
    return result
end

function basename(file)
    local sep = "/"
    if runtime.GOOS == "windows" then
        sep = "\\"
    end
    local name = string.gsub(file, "(.*" .. sep .. ")(.*)", "%2")
    return name
end
