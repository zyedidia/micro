if GetOption("formatter") == nil then
    AddOption("formatter", false)
end

MakeCommand("format", "formatter.runFormatter", 0)

function runFormatter()
    CurView():Save(false)

    local ft = CurView().Buf:FileType()
    local handle
    if ft == "go" then
        handle = io.popen("gofmt -s -w " .. CurView().Buf.Path)
    elseif ft == "fish" then
        handle = io.popen("fish_indent -w " .. CurView().Buf.Path)
    elseif ft == "shell" then
        handle = io.popen("shfmt -s -w " .. CurView().Buf.Path)
    else
        return
    end
    local result = handle:read("*a")
    handle:close()

    CurView():ReOpen()
end

function onSave(view)
    if GetOption("formatter") then
        runFormatter()
    end
end
