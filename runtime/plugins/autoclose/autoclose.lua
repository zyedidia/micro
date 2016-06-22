getmetatable('').__index = function(str,i) return string.sub(str,i,i) end

if GetOption("autoclose") == nil then
    AddOption("autoclose", true)
end

local autoclosePairs = {"\"\"", "''", "()", "{}", "[]"}
local autoNewlinePairs = {"()", "{}", "[]"}

function onRune(r)
    if not GetOption("autoclose") then
        return
    end

    local v = CurView()
    for i = 1, #autoclosePairs do
        if r == autoclosePairs[i][2] then
            local curLine = v.Buf:Line(v.Cursor.Y)

            if curLine[v.Cursor.X+1] == autoclosePairs[i][2] then
                v:Backspace()
                v:CursorRight()
                break
            end

            if v.Cursor.X > 1 and (IsWordChar(curLine[v.Cursor.X-1]) or curLine[v.Cursor.X-1] == autoclosePairs[i][1]) then
                break
            end
        end
        if r == autoclosePairs[i][1] then
            local curLine = v.Buf:Line(v.Cursor.Y)

            if v.Cursor.X == #curLine or not IsWordChar(curLine[v.Cursor.X+1]) then
                -- the '-' here is to derefence the pointer to v.Cursor.Loc which is automatically made
                -- when converting go structs to lua
                -- It needs to be dereferenced because the function expects a non pointer struct
                v.Buf:Insert(-v.Cursor.Loc, autoclosePairs[i][2])
                break
            end
        end
    end
end

function onInsertEnter()
    if not GetOption("autoclose") then
        return
    end

    local v = CurView()
    local curLine = v.Buf:Line(v.Cursor.Y)
    local lastLine = v.Buf:Line(v.Cursor.Y-1)
    local curRune = lastLine[#lastLine]
    local nextRune = curLine[1]

    for i = 1, #autoNewlinePairs do
        if curRune == autoNewlinePairs[i][1] then
            if nextRune == autoNewlinePairs[i][2] then
                v:InsertTab()
                v.Buf:Insert(-v.Cursor.Loc, "\n")
            end
        end
    end
end

function onBackspace()
    if not GetOption("autoclose") then
        return
    end

    local v = CurView()

    for i = 1, #autoclosePairs do
        local curLine = v.Buf:Line(v.Cursor.Y)
        if curLine[v.Cursor.X+1] == autoclosePairs[i][2] then
            v:Delete()
        end
    end
end
