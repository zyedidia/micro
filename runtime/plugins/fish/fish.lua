if GetOption("fishfmt") == nil then
    AddOption("fishfmt", false)
end

MakeCommand("fishfmt", "fish_fishfmt")

function fish_onSave()
    if views[mainView+1].Buf.FileType == "fish" then
        if GetOption("fishfmt") then
            fish_fishfmt()
        end
    end
end

function fish_fishfmt()
    views[mainView+1]:Save()
    local handle = io.popen("fish_indent -w " .. views[mainView+1].Buf.Path)
    local result = handle:read("*a")
    handle:close()
    
    views[mainView+1]:ReOpen()
end
