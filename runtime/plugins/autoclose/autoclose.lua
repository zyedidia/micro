function charAt(str, i)
    if i <= #str then
        return string.sub(str, i, i)
    else
        return ""
    end
end

if GetOption("autoclose") == nil then
    AddOption("autoclose", true)
end

local autoclosePairs = {"\"\"", "''", "``", "()", "{}", "[]"}
local autoNewlinePairs = {"()", "{}", "[]"}

function onRune(r, v)
    if not GetOption("autoclose") then
        return
    end

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

function preInsertNewline(v)
    if not GetOption("autoclose") then
        return
    end

    local curLine = v.Buf:Line(v.Cursor.Y)
    local curRune = charAt(curLine, v.Cursor.X)
    local nextRune = charAt(curLine, v.Cursor.X+1)
    local ws = GetLeadingWhitespace(curLine)

    for i = 1, #autoNewlinePairs do
        if curRune == charAt(autoNewlinePairs[i], 1) then
            if nextRune == charAt(autoNewlinePairs[i], 2) then
                v:InsertNewline(false)
                v:InsertTab(false)
                v.Buf:Insert(-v.Cursor.Loc, "\n" .. ws)
                return false
            end
        end
    end

    return true
end

function preBackspace(v)
    if not GetOption("autoclose") then
        return
    end

    for i = 1, #autoclosePairs do
        local curLine = v.Buf:Line(v.Cursor.Y)
        if charAt(curLine, v.Cursor.X+1) == charAt(autoclosePairs[i], 2) and charAt(curLine, v.Cursor.X) == charAt(autoclosePairs[i], 1) then
            v:Delete(false)
        end
    end

    return true
end
