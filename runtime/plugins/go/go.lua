function go_onSave()
    if view.Buf.Filetype == "Go" then
        if settings.GoImports then
            go_goimports()
        elseif settings.GoFmt then
            go_gofmt()
        end
        go_golint()
    end
end

function go_gofmt()
    local handle = io.popen("gofmt -w " .. view.Buf.Path)
    local result = handle:read("*a")
    handle:close()
    
    view:ReOpen()
end

function go_golint()
    local handle = io.popen("golint " .. view.Buf.Path)
    local result = go_split(handle:read("*a"), ":")
    handle:close()

    local file = result[1]
    local line = tonumber(result[2])
    local col = tonumber(result[3])
    local msg = result[4]

    view:ReOpen()
    view:GutterMessage(line, msg, 2)
end

function go_goimports()
    local handle = io.popen("goimports -w " .. view.Buf.Path)
    local result = go_split(handle:read("*a"), ":")
    handle:close()

    view:ReOpen()
end

function go_split(str, sep)
   local result = {}
   local regex = ("([^%s]+)"):format(sep)
   for each in str:gmatch(regex) do
      table.insert(result, each)
   end
   return result
end
