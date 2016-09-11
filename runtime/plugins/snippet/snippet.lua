local curFileType = ""
local snippets = {}

local Location = {}
Location.__index = Location

local Snippet = {}
Snippet.__index = Snippet

function Location.new(idx, ph, snip)
	local self = setmetatable({}, Location)
	self.idx = idx 
	self.ph = ph
	self.snippet = snip
    return self
end

-- offset of the location relative to the snippet start
function Location.offset(self)
	local add = 0
	for i = 1, #self.snippet.locations do
		local loc = self.snippet.locations[i]
		if loc == self then
			break
		end

		local val = loc.ph.value
		if val then
			add = add + val:len()
		end
	end
	return self.idx+add
end

function Location.startPos(self)
	local loc = self.snippet.startPos
	return loc:Move(self:offset(), self.snippet.view.buf)
end

-- returns the length of the location (but at least 1)
function Location.len(self)
	local len = 0
	if self.ph.value then
		len = self.ph.value:len()
	end
	if len <= 0 then
		len = 1
	end
	return len
end

function Location.endPos(self)
	local start = self:startPos()
	return start:Move(self:len(), self.snippet.view.buf)
end

-- check if the given loc is within the location
function Location.isWithin(self, loc)
	return loc:GreaterEqual(self:startPos()) and loc:LessEqual(self:endPos())
end

function Location.focus(self)
	local view = self.snippet.view
	local startP = self:startPos():Move(-1, view.Buf)
	local endP = self:endPos():Move(-1, view.Buf)

	while view.Cursor:LessThan(startP) do
		view.Cursor:Right()
	end
	while view.Cursor:GreaterThan(endP) do
		view.Cursor:Left()
	end

	view.Cursor:SetSelectionStart(startP)
	view.Cursor:SetSelectionEnd(endP)
end

function Location.handleInput(self, ev)
	if ev.EventType == 1 then
		if ev.Text == "\t" then
			self.snippet:focusNext()
			return true
		elseif ev.Text == "\n" then
			finishSnippet()
			return true
		else
			-- TextInput
			local offset = 1
			local sp = self:startPos()
			while sp:LessEqual(-ev.Start) do
				sp = sp:Move(1, self.snippet.view.Buf)
				offset = offset + 1
			end

			self.snippet:remove()
			local v = self.ph.value
			if v == nil then
				v = ""
			end

			self.ph.value = v:sub(0, offset-1) .. ev.Text .. v:sub(offset)
			self.snippet:insert()
		end
		return true
	elseif ev.EventType == -1 then
		local offset = 1
		local sp = self:startPos()
		while sp:LessEqual(-ev.Start) do
			sp = sp:Move(1, self.snippet.view.Buf)
			offset = offset + 1
		end

		if ev.Start.Y ~= ev.End.Y then
			return false
		end


		self.snippet:remove()

		local v = self.ph.value
		if v == nil then
			v = ""
		end

		local len = ev.End.X - ev.Start.X
		

		self.ph.value = v:sub(0, offset-1) .. v:sub(offset+len)
		self.snippet:insert()
		return true
	end

	return false
end

function Snippet.new()
	local self = setmetatable({}, Snippet)
	self.code = ""
    return self
end

function Snippet.AddCodeLine(self, line)
	if self.code ~= "" then
		self.code = self.code .. "\n"
	end
	self.code = self.code .. line
end

function Snippet.Prepare(self)
	if not self.placeholders then
		self.placeholders = {}
		self.locations = {}
		local pattern = "${(%d+):?([^}]*)}"
		while true do
  			local num, value = self.code:match(pattern)
  			if not num then
    			break
  			end
  			num = tonumber(num)
  			local idx = self.code:find(pattern)
  			self.code = self.code:gsub(pattern, "", 1)

  			local p = self.placeholders[num]
  			if not p then
  				p = {}
  				self.placeholders[num] = p
  			end
  			self.locations[#self.locations+1] = Location.new(idx, p, self)

  			if value then
  				p.value = value
  			end
		end
	end
end

function Snippet.clone(self)
	local result = Snippet.new()
	result:AddCodeLine(self.code)
	result:Prepare()
	return result
end

function Snippet.str(self)
	local res = self.code
	for i = #self.locations, 1, -1 do
		local loc = self.locations[i]
		res = res:sub(0, loc.idx-1) .. loc.ph.value .. res:sub(loc.idx)
	end
	return res
end

function Snippet.findLocation(self, loc)
	for i = 1, #self.locations do
		if self.locations[i]:isWithin(loc) then
			return self.locations[i]
		end
	end
	return nil
end

function Snippet.remove(self)
	local endPos = self.startPos:Move(self:str():len(), self.view.Buf)
	self.modText = true
	self.view.Cursor:SetSelectionStart(self.startPos)
	self.view.Cursor:SetSelectionEnd(endPos)
	self.view.Cursor:DeleteSelection()
	self.view.Cursor:ResetSelection()
	self.modText = false
end

function Snippet.insert(self)
	self.modText = true
	self.view.Buf:insert(self.startPos, self:str())
	self.modText = false
end

function Snippet.focusNext(self)
	if self.focused == nil then
		self.focused = 0
	else 
		self.focused = (self.focused + 1) % #self.placeholders
	end
	local ph = self.placeholders[self.focused+1]

	for i = 1, #self.locations do
		if self.locations[i].ph == ph then
			self.locations[i]:focus()
			return
		end
	end
end

local function CursorWord(v)
	local c = v.Cursor
	local x = c.X-1 -- start one rune before the cursor
	local result = ""
	while x >= 0 do
		local r = RuneStr(c:RuneUnder(x))
		if IsWordChar(r) then
			result = r .. result
		else
			break
		end
		x = x-1
	end

	return result
end

local function ReadSnippets(filetype)
	local snippets = {}
	local filename = JoinPaths(configDir, "plugins", "snippet", filetype .. ".snippets")

	-- first test if the file exists
	local f = io.open(filename, "r")
	if f then
		f:close()
	else
		return snippets
	end

	
	local curSnip = nil

	for line in io.lines(filename) do
		if string.match(line,"^#") then
			-- comment
		elseif line:match("^snippet") then
			curSnip = Snippet.new()
			for snipName in line:gmatch("%s(%a+)") do
				snippets[snipName] = curSnip
			end
		else
			curSnip:AddCodeLine(line:match("^\t(.*)$"))
		end
	end
	return snippets
end

local function EnsureSnippets()
	local filetype = GetOption("filetype")
	if curFileType ~= filetype then
		snippets = ReadSnippets(filetype)
		curFileType = filetype
	end
end

local currentSnippet = nil

local function finishSnippet()
	currentSnippet = nil
	-- messenger:Message("snippet editing finished")
end

function onBeforeTextEvent(ev)
	if currentSnippet ~= nil and currentSnippet.view == CurView() then
		if currentSnippet.modText then
			-- text event from the snippet. simply ignore it...
			return true
		end

		local locStart = currentSnippet:findLocation(ev.Start:Move(1, CurView().Buf))
		local locEnd = currentSnippet:findLocation(ev.End)
		if locStart ~= nil and ((locStart == locEnd) or (ev.End.Y==0 and ev.End.X==0))  then
			if locStart:handleInput(ev) then
				CurView().Cursor:Goto(-ev.C)
				return false
			end
		end
		finishSnippet()
	end

	return true
end

function foo()
	local v = CurView()
	local c = v.Cursor
	local buf = v.Buf
	local xy = Loc(c.X, c.Y)
	local name = CursorWord(v)

	EnsureSnippets()
	local curSn = snippets[name]
	if curSn then
		currentSnippet = curSn:clone()
		currentSnippet.startPos = xy:Move(-name:len(), buf)
		currentSnippet.view = v
		currentSnippet.modText = true

		c:SetSelectionStart(currentSnippet.startPos)
		c:SetSelectionEnd(xy)
		c:DeleteSelection()
		c:ResetSelection()
		
		currentSnippet.modText = false
		currentSnippet:insert()

		currentSnippet:focusNext()
	end
end

MakeCommand("foo", "snippet.foo", 0)