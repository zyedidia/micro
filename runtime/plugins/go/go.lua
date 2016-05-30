if GetOption("goimports") == nil then
    AddOption("goimports", false)
end
if GetOption("gofmt") == nil then
    AddOption("gofmt", true)
end

MakeCommand("goimports", "go_goimports")
MakeCommand("gofmt", "go_gofmt")

function go_onSave()
    if views[mainView+1].Buf.FileType == "Go" then
        if GetOption("goimports") then
            go_goimports()
        elseif GetOption("gofmt") then
            go_gofmt()
        end
    end
end

function go_gofmt()
    views[mainView+1]:Save()
    local handle = io.popen("gofmt -w " .. views[mainView+1].Buf.Path)
    local result = handle:read("*a")
    handle:close()

    views[mainView+1]:ReOpen()
end

function go_goimports()
    views[mainView+1]:Save()
    local handle = io.popen("goimports -w " .. views[mainView+1].Buf.Path)
    local result = go_split(handle:read("*a"), ":")
    handle:close()

    views[mainView+1]:ReOpen()
end

function go_split(str, sep)
    local result = {}
    local regex = ("([^%s]+)"):format(sep)
    for each in str:gmatch(regex) do
        table.insert(result, each)
    end
    return result
end
