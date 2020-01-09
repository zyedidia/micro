micro = import("micro")
buffer = import("micro/buffer")
config = import("micro/config")

function init()
    micro.SetStatusInfoFn("status.branch")
    micro.SetStatusInfoFn("status.hash")
    micro.SetStatusInfoFn("status.paste")
end

function branch(b)
    if b.Type.Kind ~= buffer.BTInfo then
        local shell = import("micro/shell")
        local strings = import("strings")

        local branch, err = shell.ExecCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
        if err == nil then
            return strings.TrimSpace(branch)
        end
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
    end
end

function paste(b)
    if config.GetGlobalOption("paste") then
        return "PASTE "
    end
    return ""
end
