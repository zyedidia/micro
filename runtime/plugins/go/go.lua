if GetOption("goimports") == nil then
    AddOption("goimports", false)
end
if GetOption("gofmt") == nil then
    AddOption("gofmt", true)
end

function go_onSave()
    if view.Buf.Filetype == "Go" then
        if GetOption("goimports") then
            go_goimports()
        elseif GetOption("gofmt") then
            go_gofmt()
        end
        go_build()
        go_golint()

        view:ReOpen()
    end
end

function go_gofmt()
    local handle = io.popen("gofmt -w " .. view.Buf.Path)
    local result = handle:read("*a")
    handle:close()
end

function go_golint()
    view:ClearGutterMessages("go-lint")

    local handle = io.popen("golint " .. view.Buf.Path)
    local lines = go_split(handle:read("*a"), "\n")
    handle:close()

    for _,line in ipairs(lines) do
        local result = go_split(line, ":")
        local line = tonumber(result[2])
        local msg = result[4]

        view:GutterMessage("go-lint", line, msg, 2)
    end
end

function go_build()
    view:ClearGutterMessages("go-build")

    local handle = io.popen("go build " .. view.Buf.Path .. " 2>&1")
    local lines = go_split(handle:read("*a"), "\n")
    handle:close()

    messenger:Message(view.Buf.Path)
    for _,line in ipairs(lines) do
        local line, msg = string.match(line, ".+:(%d+):(.+)")
        view:GutterMessage("go-build", tonumber(line), msg, 2)
    end
end

function go_goimports()
    local handle = io.popen("goimports -w " .. view.Buf.Path)
    local result = go_split(handle:read("*a"), ":")
    handle:close()
end

function go_split(str, sep)
    local result = {}
    local regex = ("([^%s]+)"):format(sep)
    for each in str:gmatch(regex) do
        table.insert(result, each)
    end
    return result
end
