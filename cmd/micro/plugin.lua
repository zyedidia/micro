go = {}

function onSave()
    if settings.GoImports then
        messenger:Message("Running goimports...")
        go.goimports()
    elseif settings.GoFmt then
        messenger:Message("Running gofmt...")
        go.gofmt()
    end
end

function go.gofmt()
    local handle = io.popen("gofmt -w " .. view.Buf.Path)
    local result = handle:read("*a")
    handle:close()

    view:ReOpen()
    messenger:Message(result)
end

function go.goimports()
    local handle = io.popen("goimports -w " .. view.Buf.Path)
    local result = handle:read("*a")
    handle:close()

    view:ReOpen()
    messenger:Message(result)
end
