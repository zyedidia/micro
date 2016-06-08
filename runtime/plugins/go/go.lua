if GetOption("goimports") == nil then
    AddOption("goimports", false)
end
if GetOption("gofmt") == nil then
    AddOption("gofmt", true)
end

MakeCommand("goimports", "go_goimports")
MakeCommand("gofmt", "go_gofmt")

function go_onSave()
    if CurView().Buf.FileType == "Go" then
        if GetOption("goimports") then
            go_goimports()
        elseif GetOption("gofmt") then
            go_gofmt()
        end
    end
end

function go_gofmt()
    CurView():Save()
    local handle = io.popen("gofmt -w " .. CurView().Buf.Path)
    local result = handle:read("*a")
    handle:close()

    CurView():ReOpen()
end

function go_goimports()
    CurView():Save()
    local handle = io.popen("goimports -w " .. CurView().Buf.Path)
    local result = go_split(handle:read("*a"), ":")
    handle:close()

    CurView():ReOpen()
end

function go_split(str, sep)
    local result = {}
    local regex = ("([^%s]+)"):format(sep)
    for each in str:gmatch(regex) do
        table.insert(result, each)
    end
    return result
end
