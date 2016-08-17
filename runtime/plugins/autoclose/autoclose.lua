function charAt(str, i)
    return string.sub(str, i, i)
end

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
        if r == charAt(autoclosePairs[i], 2) then
            local curLine = v.Buf:Line(v.Cursor.Y)

            if charAt(curLine, v.Cursor.X+1) == charAt(autoclosePairs[i], 2) then
                v:Backspace(false)
                v:CursorRight(false)
                break
            end

            if v.Cursor.X > 1 and (IsWordChar(charAt(curLine, v.Cursor.X-1)) or charAt(curLine, v.Cursor.X-1) == charAt(autoclosePairs[i], 1)) then
                break
            end
        end
        if r == charAt(autoclosePairs[i], 1) then
            local curLine = v.Buf:Line(v.Cursor.Y)

            if v.Cursor.X == #curLine or not IsWordChar(charAt(curLine, v.Cursor.X+1)) then
                -- the '-' here is to derefence the pointer to v.Cursor.Loc which is automatically made
                -- when converting go structs to lua
                -- It needs to be dereferenced because the function expects a non pointer struct
                v.Buf:Insert(-v.Cursor.Loc, charAt(autoclosePairs[i], 2))
                break
            end
        end
    end
end

function onInsertNewline()
    if not GetOption("autoclose") then
        return
    end

    local v = CurView()
    local curLine = v.Buf:Line(v.Cursor.Y)
    local lastLine = v.Buf:Line(v.Cursor.Y-1)
    local curRune = charAt(lastLine, #lastLine)
    local nextRune = charAt(curLine, 1)

    for i = 1, #autoNewlinePairs do
        if curRune == charAt(autoNewlinePairs[i], 1) then
            if nextRune == charAt(autoNewlinePairs[i], 2) then
                v:InsertTab(false)
                v.Buf:Insert(-v.Cursor.Loc, "\n")
            end
        end
    end
end

function preBackspace()
    if not GetOption("autoclose") then
        return
    end

    local v = CurView()

    for i = 1, #autoclosePairs do
        local curLine = v.Buf:Line(v.Cursor.Y)
        if charAt(curLine, v.Cursor.X+1) == charAt(autoclosePairs[i], 2) and charAt(curLine, v.Cursor.X) == charAt(autoclosePairs[i], 1) then
            v:Delete(false)
        end
    end

    return true
end
