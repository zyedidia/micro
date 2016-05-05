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

        view:ReOpen()
    end
end

function go_gofmt()
    local handle = io.popen("gofmt -w " .. view.Buf.Path)
    local result = handle:read("*a")
    handle:close()
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
