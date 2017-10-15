if GetOption("ftoptions") == nil then
    AddOption("ftoptions", true)
end

function onViewOpen(view)
    if not GetOption("ftoptions") then
        return
    end

    local ft = view.Buf.Settings["filetype"]

    if ft == "makefile" or ft == "go" then
        SetOption("tabstospaces", "off")
    elseif ft == "python" or ft == "python2" or ft == "python3" or ft == "yaml" or ft =="nim" then
        SetOption("tabstospaces", "on")
    end
end
