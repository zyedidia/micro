function onSave()
    local handle = io.popen("goimports -w view.go")
    local result = handle:read("*a")
    handle:close()

    view:ReOpen()
    messenger:Message(result)
end
