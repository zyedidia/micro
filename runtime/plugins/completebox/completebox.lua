VERSION = "1.0.0"

local micro = import("micro")
local config = import("micro/config")
local buffer = import("micro/buffer")
local overlay = import("micro/overlay")


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
			_G[full_name] = function()
				if overlay_handle then
					events[name] = true
					event_count = event_count + 1
				end
			end
		else
			_G[full_name] = function()
				if overlay_handle then
					events[name] = true
					event_count = event_count + 1
					return false
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

function event(event_name, block)
	-- Returns true if the event has occured.
	track_event(event_name, block)
	return events[event_name] or false
end

function close_overlay()
	-- Closes the overlay and untracks all events.
	untrack_events()
	overlay.DestroyOverlay(overlay_handle)
	overlay_handle = nil
end

function max_len(iter)
	-- Returns the length of the longest string in iterable
	local max = 0
	for _, item in iter do
		max = math.max(max, #item)
	end
	return max
end

function draw_autocomplete_overlay()
	local bp = micro.CurPane()
	local buf = bp.Buf

	if not buf.HasSuggestions then
		-- If there are no suggestions, we close the overlay.
		close_overlay()
		return
	end

	-- These events should not close the menu, so we track them, but
	-- we do not block them, because we want autocomplete cycling to work.
	event("CycleAutocomplete", "no_block")
	event("CycleAutocompleteBack", "no_block")

	-- Positioning adjustment - show the menu below where the cursor
	-- was when autocomplete was initiated by subtracting the length
	-- of the currently applied completion.
	local compl_len = #buf.Completions[buf.CurSuggestion+1] + 1

	-- Note: The minus dereferences the Loc pointer
	local l = -buf:GetActiveCursor().Loc
	l = overlay.BufPaneScreenLoc(bp, l)

	local x = l.X-compl_len
	local y = l.Y+1

	-- Calculate the maximum text width of the options,
	-- add 2 cells of padding
	local w = max_len(buf.Suggestions())+2

	-- Draw each option, highlight the current option
	local yoff = 0
	local style = overlay.GetColor("cursor-line")
	for i, option in buf.Suggestions() do
		local style = overlay.Style()
		if i == buf.CurSuggestion+1 then
			style = overlay.GetColor("statusline")
		end

		overlay.DrawText(" "..option, x, y+yoff, w, 1, style)
		yoff = yoff+1
	end

	reset_events()
end

function init()
	config.AddRuntimeFile("completebox", config.RTHelp, "help/completebox.md")
end

function deinit()
	close_overlay()
	untrack_events()
end

function onAutocomplete()
	if overlay_handle then return end
	reset_events()
	overlay_handle = overlay.CreateOverlay(draw_autocomplete_overlay)
end
