if GetOption("linter") == nil then
    AddOption("linter", true)
end

MakeCommand("lint", "linter.lintCommand", 0)

function lintCommand()
    CurView():Save(false)
    runLinter()
end

function runLinter()
    local ft = CurView().Buf:FileType()
    local file = CurView().Buf.Path
    local dir = DirectoryName(file)
    if OS == "windows" then
        devnull = "NUL"
    else
        devnull = "/dev/null"
    end

    if ft == "c" then
        lint("gcc", "gcc", {"-fsyntax-only", "-Wall", "-Wextra", file}, "%f:%l:%d+:.+: %m")
    elseif ft == "c++" then
        lint("gcc", "gcc", {"-fsyntax-only","-std=c++14", "-Wall", "-Wextra", file}, "%f:%l:%d+:.+: %m")
    elseif ft == "d" then
        lint("dmd", "dmd", {"-color=off", "-o-", "-w", "-wi", "-c", file}, "%f%(%l%):.+: %m")
    elseif ft == "go" then
        lint("gobuild", "go", {"build", "-o", devnull}, "%f:%l: %m")
        lint("golint", "golint", {file}, "%f:%l:%d+: %m")
    elseif ft == "java" then
        lint("javac", "javac", {"-d", dir, file}, "%f:%l: error: %m")
    elseif ft == "javascript" then
        lint("jshint", "jshint", {file}, "%f: line %l,.+, %m")
    elseif string.match(ft, "literate") then
        lint("literate", "lit", {"-c", file}, "%f:%l:%m")
    elseif ft == "lua" then
        lint("luacheck", "luacheck", {"--no-color", file}, "%f:%l:%d+: %m")
    elseif ft == "nim" then
        lint("nim", "nim", {"check", "--listFullPaths", "--stdout", "--hints:off", file}, "%f.%l, %d+. %m")
    elseif ft == "Objective-C" then
        lint("clang", "xcrun", {"clang", "-fsyntax-only", "-Wall", "-Wextra", file}, "%f:%l:%d+:.+: %m")
    elseif ft == "python" then
        lint("pyflakes", "pyflakes", {file}, "%f:%l:.-:? %m")
        lint("mypy", "mypy", {file}, "%f:%l: %m")
        lint("pylint", "pylint", {"--output-format=parseable", "--reports=no", file}, "%f:%l: %m")
    elseif ft == "shell" then
        lint("shfmt", "shfmt", {file}, "%f:%l:%d+: %m")
    elseif ft == "swift" and OS == "darwin" then
        lint("switfc", "xcrun", {"swiftc", file}, "%f:%l:%d+:.+: %m")
    elseif ft == "swift" and OS == "linux" then
        lint("switfc", "swiftc", {file}, "%f:%l:%d+:.+: %m")
    elseif ft == "yaml" then
        lint("yaml", "yamllint", {"--format", "parsable", file}, "%f:%l:%d+:.+ %m")
    end
end

function onSave(view)
    if GetOption("linter") then
        runLinter()
    else
        CurView():ClearAllGutterMessages()
    end
end

function lint(linter, cmd, args, errorformat)
    CurView():ClearGutterMessages(linter)

    JobSpawn(cmd, args, "", "", "linter.onExit", linter, errorformat)
end

function onExit(output, linter, errorformat)
    local lines = split(output, "\n")

    local regex = errorformat:gsub("%%f", "(..-)"):gsub("%%l", "(%d+)"):gsub("%%m", "(.+)")
    for _,line in ipairs(lines) do
        -- Trim whitespace
        line = line:match("^%s*(.+)%s*$")
        if string.find(line, regex) then
            local file, line, msg = string.match(line, regex)
            if basename(CurView().Buf.Path) == basename(file) then
                CurView():GutterMessage(linter, tonumber(line), msg, 2)
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
    if OS == "windows" then
        sep = "\\"
    end
    local name = string.gsub(file, "(.*" .. sep .. ")(.*)", "%2")
    return name
end
