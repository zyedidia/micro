function onViewOpen(view)
    local ft = view.Buf.Settings["filetype"]

    if ft == "makefile" or ft == "go" then
        SetOption("tabstospaces", "off")
    elseif ft == "python" then
        SetOption("tabstospaces", "on")
    end
end
