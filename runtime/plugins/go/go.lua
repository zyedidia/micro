if GetOption("goimports") == nil then
    AddOption("goimports", false)
end
if GetOption("gofmt") == nil then
    AddOption("gofmt", true)
end

MakeCommand("goimports", "go.goimports")
MakeCommand("gofmt", "go.gofmt")

function onSave()
    if CurView().Buf.FileType == "Go" then
        if GetOption("goimports") then
            goimports()
        elseif GetOption("gofmt") then
            gofmt()
        end
    end
end

function gofmt()
    CurView():Save()
    local handle = io.popen("gofmt -w " .. CurView().Buf.Path)
    local result = handle:read("*a")
    handle:close()

    CurView():ReOpen()
end

function goimports()
    CurView():Save()
    local handle = io.popen("goimports -w " .. CurView().Buf.Path)
    local result = split(handle:read("*a"), ":")
    handle:close()

    CurView():ReOpen()
end

function split(str, sep)
    local result = {}
    local regex = ("([^%s]+)"):format(sep)
    for each in str:gmatch(regex) do
        table.insert(result, each)
    end
    return result
end
