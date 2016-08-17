if GetOption("goimports") == nil then
    AddOption("goimports", false)
end
if GetOption("gofmt") == nil then
    AddOption("gofmt", true)
end

MakeCommand("goimports", "go.save", 0)
MakeCommand("gofmt", "go.save", 0)

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
    local handle = io.popen("gofmt -w " .. CurView().Buf.Path)
    local result = handle:read("*a")
    handle:close()

    CurView():ReOpen()
end

function goimports()
    local handle = io.popen("goimports -w " .. CurView().Buf.Path)
    local result = split(handle:read("*a"), ":")
    handle:close()

    CurView():ReOpen()
end

function save()
    -- This will trigger onSave and cause gofmt or goimports to run
    CurView():Save()
end

function split(str, sep)
    local result = {}
    local regex = ("([^%s]+)"):format(sep)
    for each in str:gmatch(regex) do
        table.insert(result, each)
    end
    return result
end
