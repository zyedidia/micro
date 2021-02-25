VERSION = "1.0.0"

local micro = import("micro")
local buffer = import("micro/buffer")
local config = import("micro/config")
local humanize = import("humanize")

function init()
    micro.SetStatusInfoFn("status.branch")
    micro.SetStatusInfoFn("status.hash")
    micro.SetStatusInfoFn("status.paste")
    micro.SetStatusInfoFn("status.vcol")
    micro.SetStatusInfoFn("status.lines")
    micro.SetStatusInfoFn("status.bytes")
    micro.SetStatusInfoFn("status.size")
    config.AddRuntimeFile("status", config.RTHelp, "help/status.md")
end

function lines(b)
    return tostring(b:LinesNum())
end

function vcol(b)
    return tostring(b:GetActiveCursor():GetVisualX())
end

function bytes(b)
    return tostring(b:Size())
end

function size(b)
    return humanize.Bytes(b:Size())
end

function branch(b)
    if b.Type.Kind ~= buffer.BTInfo then
        local shell = import("micro/shell")
        local strings = import("strings")

        local branch, err = shell.ExecCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
        if err == nil then
            return strings.TrimSpace(branch)
        end
        return ""
    end
end

function hash(b)
    if b.Type.Kind ~= 5 then
        local shell = import("micro/shell")
        local strings = import("strings")

        local hash, err = shell.ExecCommand("git", "rev-parse", "--short", "HEAD")
        if err == nil then
            return strings.TrimSpace(hash)
        end
        return ""
    end
end

function paste(b)
    if config.GetGlobalOption("paste") then
        return "PASTE "
    end
    return ""
end
