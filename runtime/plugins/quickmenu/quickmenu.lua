local micro = import("micro")
local config = import("micro/config")
local buffer = import("micro/buffer")
local overlay = import("micro/overlay")
local shell = import("micro/shell")
local strings = import("strings")
local pathlib = import("path")

local last_job = nil
local results = {}

function wrap_int(val, min, max)
    if min==max then return min end
    local range = max - min + 1
    return min + (val - min) % range
end

function clamp(val, min, max)
    if val < min then return min end
    if val > max then return max end
    return val
end

function array_get(arr, idx)
	if idx <= #arr then
		return arr[idx]
	else
		return nil
	end
end

function cancel_job(job)
    if job and not job.ProcessState then
        shell.JobStop(job)
    end
end

function find(query)
    cancel_job(last_job)
    local job = nil
    results = {}

    local parts = strings.Fields(query)
    local args = {".", "-type", "f"}

    for i, part in parts() do
        if i>1 then
            args[#args+1] = "-and"
        end
        args[#args+1] = "-ipath"
        args[#args+1] = "*"..part.."*"
    end

    function on_stdout(data)
        if job~=last_job then
            cancel_job(job)
            return
        end

        local new_results = strings.Split(data, "\n")
        for _, path in new_results() do
        	if #path>0 then
            	results[#results+1] = {type="file",path=path}
            end
        end

        overlay.Redraw()
        if #results>20 then
            cancel_job()
        end
    end

    function on_stderr()
        cancel_job(job)
    end

    job = shell.JobSpawn(
        "find", args, on_stdout, on_stderr, nil
    )
    last_job = job
end

function grep(query)
    cancel_job(last_job)
    local job = nil
    results = {}

    function on_stdout(data)
        if job~=last_job then
            cancel_job(job)
            return
        end

        local new_results = strings.Split(data, "\n")
        for _, res in new_results() do
            local path, line, content, ok

            path, res, ok = strings.Cut(res, ":")
            if ok then
                line, content, ok = strings.Cut(res, ":")

                if ok then
                    results[#results+1] = {
                        type="line",
                        path=path,
                        line=line,
                        content=content
                    }
                end
            end
        end

        overlay.Redraw()
        if #results>10 then
            cancel_job(job)
        end
    end

    function on_stderr()
        cancel_job(job)
    end

    job = shell.JobSpawn(
        "grep", {"-rn", query, "."},
        on_stdout, on_stderr, nil
    )
    last_job = job
end

-- Immediate-mode event handling

local overlay_handle = nil
local event_count = 0
local events = {}
local tracked_events = {}

function track_event(name, block)
    -- Registers a global handler for an event
    -- If "no_block" is passed as the second argument,
    -- the event will not be prevented.

    local full_name = "pre" .. name
    if block=="no_block" then
        full_name = "on"..name
    end

    if not tracked_events[full_name] then
        tracked_events[full_name] = true

        if block~="no_block" then
            _G[full_name] = function(...)
                if overlay_handle then
                    events[name] = {...}
                    event_count = event_count + 1
                    return false
                end
            end
        else
            _G[full_name] = function(...)
                if overlay_handle then
                    events[name] = {...}
                    event_count = event_count + 1
                end
            end
        end
    end
end

function untrack_events()
    -- Removes all global event handlers
    for e, _ in pairs(tracked_events) do
        _G[e] = nil
    end
    tracked_events = {}
end

function reset_events()
    -- Resets tracked events between redraws
    events = {}
    event_count = 0
end

function dispatch(event_name, ...)
	-- Lets us dispatch our own custom events
    local pre_event = _G["pre"..event_name]
    local on_event = _G["on"..event_name]

    if pre_event then
        local res = pre_event(...)
        if not res then
            return false
        end
    end

    if on_event then
        on_event(...)
    end
end

function event(event_name, block)
    -- Returns event arguments if the event has occurred, or nil otherwise.
    track_event(event_name, block)
    return events[event_name]
end

function close_finder()
    -- Closes the overlay and untracks all events.
    untrack_events()
    overlay.DestroyOverlay(overlay_handle)
    overlay_handle = nil
end

local mode = "quicksearch"
local query = ""
local current_result = 1

function rerun_query(query)
    if mode == "quicksearch" then
        grep(query)
    elseif mode == "quickopen" then
        find(query)
    end
end

function preRune(_, r)
	-- Note: We handle rune events like this because we could
	--       get more than one rune event per render (for example,
	--       if the redraw is slow for whatever reason and the
	--       user is typing fast).
    if overlay_handle then
        query = query .. r
        current_result = 1
        rerun_query(query)
        return false
    end
end

function draw_finder()
    local bp = micro.CurPane()

	if event("Escape") then
		close_finder()
		return
	end

	if event("Backspace") then
		query = query:sub(1, -2)
		current_result = 1
		rerun_query(query)
	end

	if event("InsertNewline") then
		local result = results[current_result]

		if result then
			local buf_path = pathlib.Clean(bp.Buf.Path)
			result.path = pathlib.Clean(result.path)

			if config.GetGlobalOption("quickmenu.newtab") and buf_path~=result.path then
				bp:NewTabCmd{result.path}
				bp = micro.CurPane()
			else
				bp:OpenCmd{result.path}
			end

			if result.type == "line" then
				bp:GotoLoc{X=0, Y=tonumber(result.line)-1}
			end
		end

		close_finder()
		return
	end

	-- TODO: Make the Left and Right arrow keys work too!
    if event("CursorUp") then current_result = current_result-1 end
    if event("CursorDown") then current_result = current_result+1 end
    local result_count = clamp(#results, 1, 10)
    current_result = wrap_int(current_result, 1, result_count+1)

    local r = overlay.BufPaneScreenRect(bp)

    local x = math.floor(r.X + r.W*0.15)
    local w = math.ceil(r.W*0.7)
    local y = r.Y + 2

    -- Draw the input box
	local input_style = overlay.GetColor("line-number")
    
    overlay.DrawRect(x-1, y, w+2, 1, input_style)
    overlay.DrawText(query, x, y, w, 1, input_style)

	if query=="" then
		if mode=="quicksearch" then
		    overlay.DrawText("Search code...", x, y, w, 1, input_style:Dim(true))
		elseif mode=="quickopen" then
			overlay.DrawText("Find file...", x, y, w, 1, input_style:Dim(true))
	    end
	end

    -- Draw the results
    local normal = overlay.GetColor("line-number")
    local highlight = overlay.GetColor("selection")

    for i, result in pairs(results) do
        local style = normal
        if i==current_result then
            style = highlight
        end

        if result.type=="line" then
            y = y+1
            overlay.DrawText(result.path..":"..result.line, x-1, y, w+2, 1, style:Bold(true))

            y = y+1
            overlay.DrawRect(x-1, y, w+2, 1, style)
            overlay.DrawText("  " .. result.content, x, y, w, 1, style)

        elseif result.type=="file" then
            y = y+1
            overlay.DrawText(result.path, x, y, w, 1, style:Bold(true))

        end

        if i>10 then break end
    end

    reset_events()
end

function open_finder(q)
    if overlay_handle then return end
    reset_events()

    if q then
        query = q
    else
        query = ""
    end

	results = {}
    current_result = 1
    overlay_handle = overlay.CreateOverlay(draw_finder)
end

function open_quickopen(_, args)
    mode = "quickopen"
    open_finder(array_get(args, 1))
end

function open_quicksearch(_, args)
    mode = "quicksearch"
    open_finder(array_get(args, 1))
end

function init()
	config.AddRuntimeFile("quickmenu", config.RTHelp, "help/quickmenu.md")
	
    config.RegisterGlobalOption("quickmenu", "newtab", true)
    config.MakeCommand("quicksearch", open_quicksearch, config.NoComplete)
    config.MakeCommand("quickopen", open_quickopen, config.NoComplete)
    config.TryBindKey("Alt-f", "command:quicksearch", false)
    config.TryBindKey("Alt-o", "command:quickopen", false)
end

function deinit()
    close_finder()
    untrack_events()
end
