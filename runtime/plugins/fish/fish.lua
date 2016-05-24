if GetOption("fishfmt") == nil then
    AddOption("fishfmt", false)
end

function fish_onSave()
    if view.Buf.FileType == "fish" then
        if GetOption("fishfmt") then
            fish_fishfmt()
        end

        view:ReOpen()
    end
end

function fish_fishfmt()
    local handle = io.popen("fish_indent -w " .. view.Buf.Path)
    local result = handle:read("*a")
    handle:close()
end